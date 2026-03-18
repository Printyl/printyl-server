package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

type JobService struct {
	mu   sync.RWMutex
	jobs map[string]*models.Job
}

func NewJobService() *JobService {
	return &JobService{
		jobs: make(map[string]*models.Job),
	}
}

func (s *JobService) AddJob() *models.Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := models.Job{
		UUID:      uuid.New(),
		CreatedAt: time.Now(),
		Status:    models.JobStatusPending,
	}
	s.jobs[job.UUID.String()] = &job

	return &job
}
