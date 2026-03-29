package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/gregor-gottschewski/printyl-server/internal/handlers"
	"github.com/gregor-gottschewski/printyl-server/internal/middleware"
	"github.com/gregor-gottschewski/printyl-server/internal/scheduler"
	"github.com/gregor-gottschewski/printyl-server/internal/service"
)

// API represents complete API structure with all API versions.
type API struct {
	mainRouter      *mux.Router
	v1              *V1
	cancelScheduler context.CancelFunc
}

type V1 struct {
	router           *mux.Router
	documentsHandler *handlers.DocumentsHandler
	statusHandler    *handlers.StatusHandler
	documentsService *service.DocumentService
	jobService       *service.JobService
}

// NewAPI creates a new API instance with all endpoints defined for all versions
func NewAPI() *API {
	ctx, cancel := context.WithCancel(context.Background())

	api := &API{
		mainRouter:      mux.NewRouter(),
		cancelScheduler: cancel,
	}

	// CORS is applied globally so that preflight OPTIONS requests are handled
	// before authentication is checked.
	api.mainRouter.Use(middleware.NewCORSMiddleware(Cfg.CORSAllowedOrigins))

	docService := service.NewDocumentService(filepath.Join(Cfg.ApplicationPath, "documents"))
	jobService := service.NewJobService()

	dockerClient, err := createDockerClient()
	if err != nil {
		cancel()
		slog.ErrorContext(ctx, "failed to create Docker client", slog.String("error", err.Error()))
		return nil
	}

	compileService := service.NewCompileService(dockerClient, jobService, Cfg.LatexImage, Cfg.ApplicationPath)
	compileScheduler := scheduler.NewCompileScheduler(jobService, compileService, 10)
	go func() {
		if err := compileScheduler.Start(ctx, 2*time.Second); err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.ErrorContext(ctx, "compile scheduler stopped", slog.String("error", err.Error()))
			}
		}
	}()

	v1 := &V1{
		router:           api.mainRouter.PathPrefix("/api/v1").Subrouter(),
		documentsHandler: &handlers.DocumentsHandler{ApplicationPath: Cfg.ApplicationPath},
		statusHandler:    handlers.NewStatusHandler(),
		documentsService: docService,
		jobService:       jobService,
	}

	v1.documentsHandler.DocumentsService = docService
	v1.documentsHandler.JobService = v1.jobService

	v1.registerDocumentsObservers()
	if err := v1.documentsService.RefreshDocuments(); err != nil {
		cancel()
		slog.ErrorContext(ctx, "Failed to initialize documents service v1", slog.String("error", err.Error()))
		return nil
	}

	if err := v1.createV1Endpoints(ctx, cancel); err != nil {
		return nil
	}

	api.v1 = v1

	return api
}

func (api *API) Start() error {
	defer api.cancelScheduler()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", Cfg.Port),
		Handler: api.mainRouter,
	}

	slog.InfoContext(context.Background(), fmt.Sprintf("Starting server on :%d", Cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		return err
	case <-quit:
		slog.InfoContext(context.Background(), "shutting down server...")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}

// createV1Endpoints registers all /api/v1 routes.
// /status is public. All /documents routes sit behind the auth middleware
// when AUTH_ENABLED=true.
func (v1 *V1) createV1Endpoints(ctx context.Context, cancel context.CancelFunc) error {
	// Public route — no authentication required.
	v1.router.HandleFunc("/status", v1.statusHandler.GetStatus).Methods("GET", "OPTIONS")

	// Protected subrouter for all document endpoints.
	protected := v1.router.PathPrefix("").Subrouter()

	if Cfg.AuthEnabled {
		authMiddleware, err := middleware.NewAuthMiddleware(ctx, Cfg.OIDCIssuerURL, Cfg.OIDCClientID)
		if err != nil {
			cancel()
			slog.ErrorContext(ctx, "failed to initialize OIDC provider", slog.String("error", err.Error()))
			return err
		}

		protected.Use(authMiddleware)
		slog.InfoContext(ctx, "auth middleware enabled", slog.String("issuer", Cfg.OIDCIssuerURL))
	} else {
		slog.WarnContext(ctx, "AUTH_ENABLED=false: all document endpoints are unprotected")
	}

	protected.HandleFunc("/documents", v1.documentsHandler.GetAllDocuments).Methods("GET", "OPTIONS")
	protected.HandleFunc("/documents/{id}/form", v1.documentsHandler.GetDocumentForm).Methods("GET", "OPTIONS")
	protected.HandleFunc("/documents/{id}/generate", v1.documentsHandler.GenerateDocument).Methods("POST", "OPTIONS")

	return nil
}

// registerDocumentsObservers registers all observers to DocumentService (v1)
// Note that DocumentService and all observers have to be initialized.
func (v1 *V1) registerDocumentsObservers() {
	v1.documentsService.AddDocumentsObserver(v1.documentsHandler)
}

func createDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}
