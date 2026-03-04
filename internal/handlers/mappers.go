package handlers

import "github.com/gregor-gottschewski/printyl-server/internal/models"

func TranslateDocument(in *models.DocumentManifest) models.Document {
	return models.Document{
		Name:        in.Title,
		Description: in.Description,
	}
}
