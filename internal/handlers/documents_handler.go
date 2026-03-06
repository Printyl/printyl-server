package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

// DocumentServicer defines the interface for document service operations.
// Decouples the handler from the service implementation.
type DocumentServicer interface {
	GetManifest(id string) (*models.DocumentManifest, error)
}

// DocumentsHandler contains DocumentService for managing documents.
type DocumentsHandler struct {
	mu               sync.RWMutex
	DocumentsService DocumentServicer
	documents        []models.Document
}

// GetAllDocuments writes a list with all documents on system to client.
func (h *DocumentsHandler) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.documents); err != nil {
		slog.ErrorContext(r.Context(), "error encoding response", slog.String("error", err.Error()))
	}
}

// OnDocumentsChanged keeps documents synchronized with the internal state.
func (h *DocumentsHandler) OnDocumentsChanged(documents map[string]*models.DocumentManifest) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.documents = nil
	for id, doc := range documents {
		h.documents = append(h.documents, TranslateDocument(doc, id))
	}
}

func (h *DocumentsHandler) GetDocumentForm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	manifest, err := h.DocumentsService.GetManifest(id)
	if err != nil {
		slog.ErrorContext(r.Context(), "error getting manifest", slog.String("id", id), slog.String("error", err.Error()))
		http.Error(w, "document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	form := models.FormResponseFromTemplate(&manifest.Template)
	if err := json.NewEncoder(w).Encode(form); err != nil {
		slog.ErrorContext(r.Context(), "error encoding response", slog.String("error", err.Error()))
	}
}
