package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

type JobService struct {
	mu   sync.Mutex
	jobs map[string]*models.Job
}

func NewJobService() *JobService {
	return &JobService{
		jobs: make(map[string]*models.Job),
	}
}

func (s *JobService) AddJob(manifest *models.DocumentManifest, generateRequest *models.GenerateRequest) *models.Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := models.Job{
		UUID:            uuid.New(),
		CreatedAt:       time.Now(),
		Status:          models.JobStatusPending,
		Manifest:        manifest,
		GenerateRequest: generateRequest,
	}
	s.jobs[job.UUID.String()] = &job

	return &job
}
