package models

// Document represents JSON object for API responses
type Document struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
