package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gregor-gottschewski/printyl-server/internal"
	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

type FormService struct {
	docID    string
	manifest *models.DocumentManifest
	genReq   *models.GenerateRequest
}

func NewFormService(docId string, manifest *models.DocumentManifest, genReq *models.GenerateRequest) *FormService {
	return &FormService{
		docID:    docId,
		manifest: manifest,
		genReq:   genReq,
	}
}

func (ds *FormService) InsertPlaceholders() error {
	file, err := ds.getLatexFile()
	if err != nil {
		return err
	}

	for key, val := range ds.genReq.Fields {
		file = strings.ReplaceAll(file, fmt.Sprintf("{$%s$}", val.Name), string(key))
	}
}

func (ds *FormService) getLatexFile() (string, error) {
	path := filepath.Join(internal.Cfg.DocumentsPath, ds.docID, "manifest.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
