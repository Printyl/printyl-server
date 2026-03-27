package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/gregor-gottschewski/printyl-server/internal/handlers"
	"github.com/gregor-gottschewski/printyl-server/internal/scheduler"
	"github.com/gregor-gottschewski/printyl-server/internal/service"
)

// API represents complete API structure with all API versions.
type API struct {
	mainRouter *mux.Router
	v1         *V1
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
	ctx := context.Background()

	api := &API{
		mainRouter: mux.NewRouter(),
	}

	docService := service.NewDocumentService(filepath.Join(Cfg.ApplicationPath, "documents"))
	jobService := service.NewJobService()

	dockerClient, err := createDockerClient()
	if err != nil {
		slog.ErrorContext(ctx, "failed to create Docker client", slog.String("error", err.Error()))
		return nil
	}

	compileService := service.NewCompileService(dockerClient, jobService, Cfg.LatexImage, Cfg.ApplicationPath)
	compileScheduler := scheduler.NewCompileScheduler(jobService, compileService, 10)
	go func() {
		if err := compileScheduler.Start(ctx, 2*time.Second); err != nil {
			slog.ErrorContext(ctx, "compile scheduler stopped", slog.String("error", err.Error()))
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
		slog.ErrorContext(ctx, "Failed to initialize documents service v1", slog.String("error", err.Error()))
		return nil
	}

	v1.createV1Endpoints()

	api.v1 = v1

	return api
}

func (api *API) Start() error {
	slog.InfoContext(context.Background(), fmt.Sprintf("Starting server on :%d", Cfg.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", Cfg.Port), api.mainRouter)
}

func (v1 *V1) createV1Endpoints() {
	v1.router.HandleFunc("/status", v1.statusHandler.GetStatus).Methods("GET")
	v1.router.HandleFunc("/documents", v1.documentsHandler.GetAllDocuments).Methods("GET")
	v1.router.HandleFunc("/documents/{id}/form", v1.documentsHandler.GetDocumentForm).Methods("GET")
	v1.router.HandleFunc("/documents/{id}/generate", v1.documentsHandler.GenerateDocument).Methods("POST")
}

// registerDocumentsObservers registers all observers to DocumentService (v1)
// Note that DocumentService and all observers have to be initialized.
func (v1 *V1) registerDocumentsObservers() {
	v1.documentsService.AddDocumentsObserver(v1.documentsHandler)
}

func createDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}
