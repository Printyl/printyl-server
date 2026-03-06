package service

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/gregor-gottschewski/printyl-server/internal/models"
	"gopkg.in/yaml.v3"
)

// DocumentService is a central element to handle document related actions.
// Note that the models.DocumentManifest can have another title than the ID of the document.
// Title defined in models.DocumentManifest is not unique!
type DocumentService struct {
	mu            sync.RWMutex
	documentsPath string
	Documents     map[string]*models.DocumentManifest
	docObservers  []models.DocumentObserver
}

// NewDocumentService needs a path to the Documents directory.
func NewDocumentService(dPath string) *DocumentService {
	return &DocumentService{
		Documents:     make(map[string]*models.DocumentManifest),
		documentsPath: dPath,
	}
}

// RefreshDocuments reloads all documents in documentsPath into Documents.
func (ds *DocumentService) RefreshDocuments() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	dir, err := os.ReadDir(ds.documentsPath)
	if err != nil {
		slog.Error("could not read documents directory", slog.String("error", err.Error()))
		return err
	}

	ds.Documents = ds.readStore(dir)

	ds.notify()

	return nil
}

// readStore reads all documents in store and loads them to Documents.
func (ds *DocumentService) readStore(documents []os.DirEntry) map[string]*models.DocumentManifest {
	docs := make(map[string]*models.DocumentManifest)

	for _, entry := range documents {
		if !entry.IsDir() {
			continue
		}

		doc, err := readDocumentManifest(ds.documentsPath, entry.Name())
		if err != nil {
			slog.Error("could not load document", slog.String("document", entry.Name()), slog.String("error", err.Error()))
			continue
		}
		docs[entry.Name()] = doc
	}

	return docs
}

func (ds *DocumentService) AddDocumentsObserver(observer models.DocumentObserver) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.docObservers = append(ds.docObservers, observer)
}

// readDocumentManifest loads one document to the documentsPath by its entry name
func readDocumentManifest(storePath string, entry string) (*models.DocumentManifest, error) {
	path := filepath.Join(storePath, entry, "manifest.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("could not read manifest file", slog.String("path", path), slog.String("error", err.Error()))
		return &models.DocumentManifest{}, err
	}

	var doc models.DocumentManifest
	err = yaml.Unmarshal(data, &doc)
	if err != nil {
		slog.Error("could not unmarshal YAML", slog.String("path", path), slog.String("error", err.Error()))
		return &models.DocumentManifest{}, err
	}

	return &doc, nil
}

// notify document observers
func (ds *DocumentService) notify() {
	for _, observer := range ds.docObservers {
		observer.OnDocumentsChanged(ds.Documents)
	}
}

// GetManifest returns the document manifest file by parsing <application-root>/documents/<id>/manifest.yaml
func (ds *DocumentService) GetManifest(id string) (*models.DocumentManifest, error) {
	return readDocumentManifest(ds.documentsPath, id)
}
