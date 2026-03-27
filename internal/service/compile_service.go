package service

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

// JobStatusUpdater is used by CompileService to update job status.
type JobStatusUpdater interface {
	SetStatus(uuid uuid.UUID, status models.JobStatus)
}

// CompileService compiles a LaTeX job by running pdflatex inside a
// dedicated Docker container per job.
type CompileService struct {
	docker          *client.Client
	jobUpdater      JobStatusUpdater
	imageName       string
	applicationPath string
	containerConfig *container.Config
}

// NewCompileService creates a new CompileService.
func NewCompileService(docker *client.Client, jobUpdater JobStatusUpdater, imageName string, applicationPath string) *CompileService {
	cs := &CompileService{
		docker:          docker,
		jobUpdater:      jobUpdater,
		imageName:       imageName,
		applicationPath: applicationPath,
	}

	cs.containerConfig = &container.Config{
		Image:      cs.imageName,
		Cmd:        []string{"pdflatex", "-interaction=nonstopmode", "-no-shell-escape", "-halt-on-error", "out.tex"},
		WorkingDir: "/jobs",
	}

	return cs
}

// Compile starts a new Docker container for the given job, runs pdflatex on
// the prepared out.tex file, waits for the result, and updates the job status
// accordingly. The container will be deleted after compile.
func (c *CompileService) Compile(ctx context.Context, job models.Job) error {
	c.jobUpdater.SetStatus(job.UUID, models.JobStatusRunning)

	jobDir, err := filepath.Abs(filepath.Join(c.applicationPath, "jobs", job.UUID.String()))
	if err != nil {
		c.jobUpdater.SetStatus(job.UUID, models.JobStatusFailed)
		return fmt.Errorf("get job dir: %w", err)
	}

	resp, err := c.docker.ContainerCreate(
		ctx,
		c.containerConfig,
		c.hostConfig(jobDir),
		nil,
		nil,
		"",
	)
	if err != nil {
		c.jobUpdater.SetStatus(job.UUID, models.JobStatusFailed)
		return fmt.Errorf("creating container for job %s: %w", job.UUID, err)
	}

	defer func() {
		if removeErr := c.docker.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true}); removeErr != nil {
			slog.WarnContext(ctx, "failed to remove compile container",
				slog.String("container", resp.ID),
				slog.String("job", job.UUID.String()),
				slog.String("error", removeErr.Error()),
			)
		}
	}()

	if err := c.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		c.jobUpdater.SetStatus(job.UUID, models.JobStatusFailed)
		return fmt.Errorf("starting container for job %s: %w", job.UUID, err)
	}

	statusCh, errCh := c.docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case waitErr := <-errCh:
		if waitErr != nil {
			c.jobUpdater.SetStatus(job.UUID, models.JobStatusFailed)
			return fmt.Errorf("waiting for container of job %s: %w", job.UUID, waitErr)
		}
	case result := <-statusCh:
		if result.Error != nil {
			c.jobUpdater.SetStatus(job.UUID, models.JobStatusFailed)
			return fmt.Errorf("container error for job %s: %s", job.UUID, result.Error.Message)
		}
		if result.StatusCode != 0 {
			c.jobUpdater.SetStatus(job.UUID, models.JobStatusFailed)
			return fmt.Errorf("pdflatex exited with code %d for job %s", result.StatusCode, job.UUID)
		}
	}

	c.jobUpdater.SetStatus(job.UUID, models.JobStatusCompleted)
	return nil
}

func (c *CompileService) hostConfig(jobDir string) *container.HostConfig {
	return &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: jobDir,
				Target: "/jobs",
			},
		},
		AutoRemove:     false,
		NetworkMode:    "none",                             // no network access needed for pdflatex
		ReadonlyRootfs: true,                               // container filesystem is read-only (job dir is the only writable mount)
		CapDrop:        []string{"ALL"},                    // drop every Linux capability
		SecurityOpt:    []string{"no-new-privileges:true"}, // prevent privilege escalation
	}
}
