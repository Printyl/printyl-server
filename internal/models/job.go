package models

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusFailed    = "failed"
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
)

type Job struct {
	UUID      uuid.UUID
	CreatedAt time.Time
	Status    JobStatus
}

type JobResponse struct {
	UUID uuid.UUID `json:"uuid"`
}
