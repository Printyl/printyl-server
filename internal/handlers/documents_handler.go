package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

// DocumentsHandler contains DocumentService for managing documents.
type DocumentsHandler struct {
	documents []models.Document
}

// GetAllDocuments writes a list with all documents on system to client.
func (h *DocumentsHandler) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.documents); err != nil {
		slog.ErrorContext(r.Context(), "error encoding response", slog.String("error", err.Error()))
	}
}

// OnDocumentsChanged keeps documents synchronized with the internal state.
func (h *DocumentsHandler) OnDocumentsChanged(documents map[string]*models.DocumentManifest) {
	h.documents = nil
	for id, doc := range documents {
		h.documents = append(h.documents, TranslateDocument(doc, id))
	}
}
