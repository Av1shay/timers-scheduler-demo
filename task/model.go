package task

import (
	"fmt"
	"time"
)

type Task struct {
	ID         int       `json:"id"`
	WebhookURL string    `json:"webhookUrl"`
	DueDate    time.Time `json:"dueDate"`
}

type ApiError struct {
	Code          int
	Message       string
	ClientMessage string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}
