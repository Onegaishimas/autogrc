package push

import (
	"time"

	"github.com/google/uuid"
)

// StartPushRequest is the request to start a push job.
type StartPushRequest struct {
	StatementIDs []uuid.UUID `json:"statement_ids"`
}

// StartPushResponse is the response after starting a push job.
type StartPushResponse struct {
	Job JobResponse `json:"job"`
}

// JobResponse represents a push job in API responses.
type JobResponse struct {
	ID          uuid.UUID              `json:"id"`
	Status      string                 `json:"status"`
	TotalCount  int                    `json:"total_count"`
	Completed   int                    `json:"completed"`
	Succeeded   int                    `json:"succeeded"`
	Failed      int                    `json:"failed"`
	Results     []StatementResultResp  `json:"results"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// StatementResultResp represents a push result for a single statement.
type StatementResultResp struct {
	StatementID uuid.UUID  `json:"statement_id"`
	Success     bool       `json:"success"`
	Error       *string    `json:"error,omitempty"`
	PushedAt    *time.Time `json:"pushed_at,omitempty"`
}

// PushStatusResponse is the response for getting push job status.
type PushStatusResponse struct {
	Job JobResponse `json:"job"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
