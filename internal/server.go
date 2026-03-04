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
}

// NewAPI creates a new API instance with all endpoints defined for all versions
func NewAPI() *API {
	api := &API{
		mainRouter: mux.NewRouter(),
	}

	v1 := &V1{
		router:           api.mainRouter.PathPrefix("/api/v1").Subrouter(),
		documentsHandler: &handlers.DocumentsHandler{},
		statusHandler:    handlers.NewStatusHandler(),
	}

	if err := v1.initDocumentsServiceV1(); err != nil {
		slog.ErrorContext(context.Background(), "Failed to initialize documents service v1: %v", err)
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
}

func (v1 *V1) initDocumentsServiceV1() error {
	v1.documentsService = service.NewDocumentService(Cfg.DocumentsPath)
	v1.registrationDocumentsObservers()
	if err := v1.documentsService.RefreshDocuments(); err != nil {
		return err
	}
	return nil
}

func (v1 *V1) registrationDocumentsObservers() {
	v1.documentsService.AddDocumentsObserver(v1.documentsHandler)
}
