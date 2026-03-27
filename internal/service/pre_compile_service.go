package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

type PreCompileService struct {
	Job             models.Job
	applicationPath string
}

func NewPreCompileService(job models.Job, applicationPath string) *PreCompileService {
	return &PreCompileService{
		Job:             job,
		applicationPath: applicationPath,
	}
}

func (p *PreCompileService) CreateTempCompileDirectory() error {
	tmp := filepath.Join(p.applicationPath, "jobs", p.Job.UUID.String())
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}
	return nil
}

// CopyTemplate copies the given template file (path) to the job directory using streams.
func (p *PreCompileService) CopyTemplate(src string) error {
	dst := filepath.Join(p.applicationPath, "jobs", p.Job.UUID.String(), p.Job.Manifest.TexFile)
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, fin)
	if err != nil {
		return err
	}
	return nil
}

// InsertPlaceholder inserts given placeholders from the GenerateRequest into the specified template.
func (p *PreCompileService) InsertPlaceholder() error {
	src := filepath.Join(p.applicationPath, "jobs", p.Job.UUID.String(), p.Job.Manifest.TexFile)

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dst := filepath.Join(p.applicationPath, "jobs", p.Job.UUID.String(), "out.tex")

	var out *os.File
	out, err = os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if p.Job.GenerateRequest.Fields != nil {
			for _, field := range *p.Job.GenerateRequest.Fields {
				if len(field.Value) == 0 {
					line = strings.ReplaceAll(line, fmt.Sprintf("{{%s}}", field.Name), "")
					continue
				}
				line = strings.ReplaceAll(line, fmt.Sprintf("{{%s}}", field.Name), field.Value)
			}
		}
		_, err := out.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
