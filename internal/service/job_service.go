package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gregor-gottschewski/printyl-server/internal/models"
)

type JobService struct {
	mu       sync.Mutex
	jobs     map[string]*models.Job
	jobQueue *[]string
}

func NewJobService() *JobService {
	return &JobService{
		jobs:     make(map[string]*models.Job),
		jobQueue: new([]string),
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
	*s.jobQueue = append(*s.jobQueue, job.UUID.String())

	return &job
}

func (s *JobService) SetStatus(uuid uuid.UUID, status models.JobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[uuid.String()]
	if !ok {
		return
	}

	job.Status = status
}

// Dequeue pops n elements from the queue as copy
func (s *JobService) Dequeue(elements int) []models.Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	limit := elements
	if limit > len(*s.jobQueue) {
		limit = len(*s.jobQueue)
	}

	ids := (*s.jobQueue)[:limit]
	*s.jobQueue = (*s.jobQueue)[limit:]

	jobs := make([]models.Job, len(ids))
	for i, id := range ids {
		jobs[i] = *s.jobs[id]
	}

	return jobs
}
