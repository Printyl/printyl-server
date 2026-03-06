package handlers

import "github.com/gregor-gottschewski/printyl-server/internal/models"

// TranslateDocument converts a models.DocumentManifest to a models.Document.
func TranslateDocument(in *models.DocumentManifest, id string) models.Document {
	return models.Document{
		ID:          id,
		Name:        in.Title,
		Description: in.Description,
	}
}
