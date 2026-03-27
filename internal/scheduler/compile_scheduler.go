package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/gregor-gottschewski/printyl-server/internal/models"
	"github.com/gregor-gottschewski/printyl-server/internal/service"
)

// CompileScheduler is used to run service.CompileService in specified interval.
type CompileScheduler struct {
	sem            chan struct{}
	jobService     *service.JobService
	compileService *service.CompileService
}

func NewCompileScheduler(js *service.JobService, cs *service.CompileService, queueSize int) *CompileScheduler {
	return &CompileScheduler{
		sem:            make(chan struct{}, queueSize),
		jobService:     js,
		compileService: cs,
	}
}

// Start sets a ticker to run compile services in specified interval.
func (s *CompileScheduler) Start(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.run(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// run runs all jobs in queue.
func (s *CompileScheduler) run(ctx context.Context) {
	available := cap(s.sem) - len(s.sem)
	if available == 0 {
		return
	}

	jobs := s.jobService.Dequeue(available)
	for _, job := range jobs {
		s.sem <- struct{}{}
		go func(j models.Job) {
			defer func() { <-s.sem }()
			if err := s.compileService.Compile(ctx, j); err != nil {
				slog.ErrorContext(ctx, "compile job failed",
					slog.String("job", j.UUID.String()),
					slog.String("error", err.Error()),
				)
			}
		}(job)
	}
}
