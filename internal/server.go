package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gregor-gottschewski/printyl-server/internal/handlers"
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
	api := &API{
		mainRouter: mux.NewRouter(),
	}

	api.mainRouter.Use(corsMiddleware)

	docService := service.NewDocumentService(Cfg.DocumentsPath)

	v1 := &V1{
		router:           api.mainRouter.PathPrefix("/api/v1").Subrouter(),
		documentsHandler: &handlers.DocumentsHandler{},
		statusHandler:    handlers.NewStatusHandler(),
		documentsService: docService,
		jobService:       service.NewJobService(),
	}

	v1.documentsHandler.DocumentsService = docService
	v1.documentsHandler.JobService = v1.jobService

	v1.registerDocumentsObservers()
	if err := v1.documentsService.RefreshDocuments(); err != nil {
		slog.ErrorContext(context.Background(), "Failed to initialize documents service v1", slog.String("error", err.Error()))
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

// corsMiddleware adds CORS headers and handles preflight requests.
// Allows all! Not production safe.
func corsMiddleware(next http.Handler) http.Handler {
	// TODO: Remove for production safe environment
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
