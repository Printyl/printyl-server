package models

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusFailed    JobStatus = "failed"
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
)

type Job struct {
	UUID      uuid.UUID
	CreatedAt time.Time
	Status    JobStatus
}

type JobResponse struct {
	UUID uuid.UUID `json:"uuid"`
}
