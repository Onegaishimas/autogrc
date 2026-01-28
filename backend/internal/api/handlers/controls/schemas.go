package controls

import "github.com/controlcrud/backend/internal/domain/controls"

// PolicyStatementDTO represents a policy statement in API responses.
type PolicyStatementDTO struct {
	ID               string `json:"id"`
	Number           string `json:"number"`
	Name             string `json:"name"`
	ShortDescription string `json:"short_description"`
	Description      string `json:"description,omitempty"`
	State            string `json:"state"`
	Category         string `json:"category,omitempty"`
	ControlFamily    string `json:"control_family,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// PaginationDTO represents pagination info in API responses.
type PaginationDTO struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// ListPolicyStatementsResponse represents the response for listing policy statements.
type ListPolicyStatementsResponse struct {
	Items      []PolicyStatementDTO `json:"items"`
	Pagination PaginationDTO        `json:"pagination"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NewPolicyStatementDTO creates a DTO from a domain model.
func NewPolicyStatementDTO(ps *controls.PolicyStatement) PolicyStatementDTO {
	dto := PolicyStatementDTO{
		ID:               ps.ID,
		Number:           ps.Number,
		Name:             ps.Name,
		ShortDescription: ps.ShortDescription,
		Description:      ps.Description,
		State:            ps.State,
		Category:         ps.Category,
		ControlFamily:    ps.ControlFamily,
	}

	if !ps.CreatedAt.IsZero() {
		dto.CreatedAt = ps.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	if !ps.UpdatedAt.IsZero() {
		dto.UpdatedAt = ps.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	return dto
}

// NewListPolicyStatementsResponse creates a response from domain list result.
func NewListPolicyStatementsResponse(result *controls.ListResult) *ListPolicyStatementsResponse {
	items := make([]PolicyStatementDTO, len(result.Items))
	for i, item := range result.Items {
		items[i] = NewPolicyStatementDTO(&item)
	}

	return &ListPolicyStatementsResponse{
		Items: items,
		Pagination: PaginationDTO{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalCount: result.TotalCount,
			TotalPages: result.TotalPages,
		},
	}
}
