package connection

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/controlcrud/backend/internal/domain/connection"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for connection management.
type Handler struct {
	service *connection.Service
}

// NewHandler creates a new connection handler.
func NewHandler(service *connection.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the connection routes with the provided mux.
// All routes are prefixed with /api/v1/connection
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/connection/status", h.GetStatus)
	mux.HandleFunc("POST /api/v1/connection/config", h.SaveConfig)
	mux.HandleFunc("POST /api/v1/connection/test", h.TestConnection)
	mux.HandleFunc("DELETE /api/v1/connection", h.DeleteConnection)
}

// GetStatus handles GET /api/v1/connection/status
// Returns the current ServiceNow connection status.
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status, err := h.service.GetStatus(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve connection status")
		return
	}

	writeJSON(w, http.StatusOK, NewStatusResponse(status))
}

// SaveConfig handles POST /api/v1/connection/config
// Saves or updates the ServiceNow connection configuration.
func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON in request body")
		return
	}

	// Validate request
	if err := validateConfigRequest(&req); err != nil {
		writeValidationError(w, err)
		return
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if uid, ok := ctx.Value("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	// Save configuration
	conn, err := h.service.SaveConfig(ctx, req.ToConfigInput(), userID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &ConfigResponse{
		ID:          conn.ID.String(),
		InstanceURL: conn.InstanceURL,
		AuthMethod:  string(conn.AuthMethod),
		Status:      string(conn.LastTestStatus),
		Message:     "Configuration saved successfully",
	})
}

// TestConnection handles POST /api/v1/connection/test
// Tests the current ServiceNow connection.
func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.service.TestConnection(ctx)
	if err != nil {
		if errors.Is(err, connection.ErrConnectionNotFound) {
			writeError(w, http.StatusNotFound, "not_configured", "No connection configured. Please save configuration first.")
			return
		}
		// Return the result with error details from ServiceNow test
		if result != nil && result.ErrorMessage != "" {
			writeJSON(w, http.StatusOK, NewTestResponse(result))
			return
		}
		writeError(w, http.StatusInternalServerError, "test_failed", "Failed to test connection")
		return
	}

	writeJSON(w, http.StatusOK, NewTestResponse(result))
}

// DeleteConnection handles DELETE /api/v1/connection
// Deletes the current ServiceNow connection configuration.
func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.service.DeleteConnection(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", "Failed to delete connection")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Connection deleted successfully",
	})
}

// validateConfigRequest validates the configuration request.
func validateConfigRequest(req *ConfigRequest) error {
	var validationErrors []ValidationError

	if req.InstanceURL == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "instance_url",
			Message: "Instance URL is required",
		})
	}

	if req.AuthMethod == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "auth_method",
			Message: "Authentication method is required",
		})
	} else if req.AuthMethod != "basic" && req.AuthMethod != "oauth" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "auth_method",
			Message: "Authentication method must be 'basic' or 'oauth'",
		})
	}

	if req.AuthMethod == "basic" {
		if req.Username == "" {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "username",
				Message: "Username is required for basic authentication",
			})
		}
		if req.Password == "" {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "password",
				Message: "Password is required for basic authentication",
			})
		}
	}

	if req.AuthMethod == "oauth" {
		if req.OAuthClientID == "" {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "oauth_client_id",
				Message: "OAuth Client ID is required for OAuth authentication",
			})
		}
		if req.OAuthClientSecret == "" {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "oauth_client_secret",
				Message: "OAuth Client Secret is required for OAuth authentication",
			})
		}
		if req.OAuthTokenURL == "" {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "oauth_token_url",
				Message: "OAuth Token URL is required for OAuth authentication",
			})
		}
	}

	if len(validationErrors) > 0 {
		return &validationErrorList{errors: validationErrors}
	}

	return nil
}

type validationErrorList struct {
	errors []ValidationError
}

func (v *validationErrorList) Error() string {
	return "validation failed"
}

// handleDomainError converts domain errors to HTTP responses.
func handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, connection.ErrInstanceURLRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "instance_url", Message: "Instance URL is required"}},
		})
	case errors.Is(err, connection.ErrAuthMethodRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "auth_method", Message: "Authentication method is required"}},
		})
	case errors.Is(err, connection.ErrInvalidAuthMethod):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "auth_method", Message: "Invalid authentication method"}},
		})
	case errors.Is(err, connection.ErrUsernameRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "username", Message: "Username is required"}},
		})
	case errors.Is(err, connection.ErrPasswordRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "password", Message: "Password is required"}},
		})
	case errors.Is(err, connection.ErrClientIDRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "oauth_client_id", Message: "OAuth Client ID is required"}},
		})
	case errors.Is(err, connection.ErrClientSecretRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "oauth_client_secret", Message: "OAuth Client Secret is required"}},
		})
	case errors.Is(err, connection.ErrTokenURLRequired):
		writeValidationError(w, &validationErrorList{
			errors: []ValidationError{{Field: "oauth_token_url", Message: "OAuth Token URL is required"}},
		})
	case errors.Is(err, connection.ErrConnectionNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Connection not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred")
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, errorCode, message string) {
	writeJSON(w, status, &ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// writeValidationError writes a validation error response.
func writeValidationError(w http.ResponseWriter, err error) {
	if ve, ok := err.(*validationErrorList); ok {
		writeJSON(w, http.StatusBadRequest, &ValidationErrorResponse{
			Error:   "validation_error",
			Message: "Request validation failed",
			Fields:  ve.errors,
		})
		return
	}
	writeError(w, http.StatusBadRequest, "validation_error", err.Error())
}
