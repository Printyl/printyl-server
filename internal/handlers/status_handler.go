package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type StatusHandler struct {
	status string
}

type StatusResponse struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

func NewStatusHandler() *StatusHandler {
	return &StatusHandler{status: "ok"}
}

func (s *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := StatusResponse{
		Name:      s.status,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}
