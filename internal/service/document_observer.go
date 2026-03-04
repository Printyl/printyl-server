package service

import "github.com/gregor-gottschewski/printyl-server/internal/models"

// DocumentObserver gives the possibility for other implementations to observe changes in documents.
type DocumentObserver interface {
	OnDocumentsChanged(documents map[string]*models.DocumentManifest)
}
