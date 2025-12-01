// internal/audit/interface/http/v1/requests.go
package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/hexagonal-go/internal/audit/application/dto"
)

// ============================================================================
// Request Parsers
// ============================================================================

// ParseGetEntryRequest parses a get entry request from URL path.
func ParseGetEntryRequest(r *http.Request) (dto.GetEntryRequest, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return dto.GetEntryRequest{}, fmt.Errorf("entry ID is required")
	}

	return dto.GetEntryRequest{ID: id}, nil
}

// ParseListEntriesRequest parses query parameters for listing audit entries.
func ParseListEntriesRequest(r *http.Request) (dto.ListEntriesRequest, error) {
	req := dto.ListEntriesRequest{
		Limit:  50,
		Offset: 0,
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return req, fmt.Errorf("invalid limit parameter")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return req, fmt.Errorf("invalid offset parameter")
		}
		req.Offset = offset
	}

	// Parse filters
	req.TenantID = r.URL.Query().Get("tenant_id")
	req.UserID = r.URL.Query().Get("user_id")
	req.EventType = r.URL.Query().Get("event_type")
	req.Source = r.URL.Query().Get("source")
	req.CorrelationID = r.URL.Query().Get("correlation_id")

	// Parse timestamp filters
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return req, fmt.Errorf("invalid 'from' timestamp format, use RFC3339")
		}
		req.FromTimestamp = &t
	}

	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return req, fmt.Errorf("invalid 'to' timestamp format, use RFC3339")
		}
		req.ToTimestamp = &t
	}

	// Validate
	if err := validateListEntriesRequest(req); err != nil {
		return req, err
	}

	return req, nil
}

// ParseGetResourceTrailRequest parses a resource trail request.
func ParseGetResourceTrailRequest(r *http.Request) (dto.GetResourceTrailRequest, error) {
	resourceType := chi.URLParam(r, "type")
	resourceID := chi.URLParam(r, "id")

	if resourceType == "" {
		return dto.GetResourceTrailRequest{}, fmt.Errorf("resource type is required")
	}
	if resourceID == "" {
		return dto.GetResourceTrailRequest{}, fmt.Errorf("resource ID is required")
	}

	req := dto.GetResourceTrailRequest{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     r.URL.Query().Get("tenant_id"),
		Limit:        50,
		Offset:       0,
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return req, fmt.Errorf("invalid limit parameter")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return req, fmt.Errorf("invalid offset parameter")
		}
		req.Offset = offset
	}

	// Parse timestamp filters
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return req, fmt.Errorf("invalid 'from' timestamp format, use RFC3339")
		}
		req.FromTimestamp = &t
	}

	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return req, fmt.Errorf("invalid 'to' timestamp format, use RFC3339")
		}
		req.ToTimestamp = &t
	}

	// Validate
	if err := validatePaginationRequest(req.Limit, req.Offset); err != nil {
		return req, err
	}

	return req, nil
}

// ParseGetActorActivityRequest parses an actor activity request.
func ParseGetActorActivityRequest(r *http.Request) (dto.GetActorActivityRequest, error) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return dto.GetActorActivityRequest{}, fmt.Errorf("user ID is required")
	}

	req := dto.GetActorActivityRequest{
		UserID:    userID,
		TenantID:  r.URL.Query().Get("tenant_id"),
		EventType: r.URL.Query().Get("event_type"),
		Source:    r.URL.Query().Get("source"),
		Limit:     50,
		Offset:    0,
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return req, fmt.Errorf("invalid limit parameter")
		}
		req.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return req, fmt.Errorf("invalid offset parameter")
		}
		req.Offset = offset
	}

	// Parse timestamp filters
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return req, fmt.Errorf("invalid 'from' timestamp format, use RFC3339")
		}
		req.FromTimestamp = &t
	}

	if to := r.URL.Query().Get("to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return req, fmt.Errorf("invalid 'to' timestamp format, use RFC3339")
		}
		req.ToTimestamp = &t
	}

	// Validate
	if err := validatePaginationRequest(req.Limit, req.Offset); err != nil {
		return req, err
	}

	return req, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateListEntriesRequest(req dto.ListEntriesRequest) error {
	return validatePaginationRequest(req.Limit, req.Offset)
}

func validatePaginationRequest(limit, offset int) error {
	if limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}
