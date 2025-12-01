// internal/audit/application/dto/mappers.go
package dto

import (
	"github.com/0xsj/hexagonal-go/internal/audit/domain"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ============================================================================
// Domain -> DTO Mappers
// ============================================================================

// EntryToDTO converts a domain AuditEntry to AuditEntryDTO.
func EntryToDTO(entry *domain.AuditEntry) AuditEntryDTO {
	return AuditEntryDTO{
		ID:            entry.ID().String(),
		EventID:       entry.EventID(),
		EventType:     entry.EventType(),
		Source:        entry.Source(),
		Timestamp:     entry.Timestamp(),
		TenantID:      entry.TenantID(),
		UserID:        entry.UserID(),
		CorrelationID: entry.CorrelationID(),
		Payload:       entry.Payload(),
		Metadata:      entry.Metadata(),
		CreatedAt:     entry.CreatedAt(),
	}
}

// EntryToSummaryDTO converts a domain AuditEntry to AuditEntrySummaryDTO.
func EntryToSummaryDTO(entry *domain.AuditEntry) AuditEntrySummaryDTO {
	return AuditEntrySummaryDTO{
		ID:            entry.ID().String(),
		EventType:     entry.EventType(),
		Source:        entry.Source(),
		Timestamp:     entry.Timestamp(),
		TenantID:      entry.TenantID(),
		UserID:        entry.UserID(),
		CorrelationID: entry.CorrelationID(),
	}
}

// EntriesToDTOs converts a slice of domain AuditEntries to AuditEntryDTOs.
func EntriesToDTOs(entries []*domain.AuditEntry) []AuditEntryDTO {
	dtos := make([]AuditEntryDTO, len(entries))
	for i, entry := range entries {
		dtos[i] = EntryToDTO(entry)
	}
	return dtos
}

// EntriesToSummaryDTOs converts a slice of domain AuditEntries to AuditEntrySummaryDTOs.
func EntriesToSummaryDTOs(entries []*domain.AuditEntry) []AuditEntrySummaryDTO {
	dtos := make([]AuditEntrySummaryDTO, len(entries))
	for i, entry := range entries {
		dtos[i] = EntryToSummaryDTO(entry)
	}
	return dtos
}

// ============================================================================
// Request -> Domain Mappers
// ============================================================================

// MapFiltersFromListRequest converts a ListEntriesRequest to domain.Filters.
func MapFiltersFromListRequest(req ListEntriesRequest) domain.Filters {
	filters := domain.Filters{}

	if req.TenantID != "" {
		filters.TenantID = &req.TenantID
	}
	if req.UserID != "" {
		filters.UserID = &req.UserID
	}
	if req.EventType != "" {
		filters.EventType = &req.EventType
	}
	if req.Source != "" {
		filters.Source = &req.Source
	}
	if req.CorrelationID != "" {
		filters.CorrelationID = &req.CorrelationID
	}
	if req.FromTimestamp != nil {
		ts := types.NewTimestamp(*req.FromTimestamp)
		filters.FromTimestamp = &ts
	}
	if req.ToTimestamp != nil {
		ts := types.NewTimestamp(*req.ToTimestamp)
		filters.ToTimestamp = &ts
	}

	return filters
}

// MapPageFromRequest extracts pagination from a request.
func MapPageFromRequest(limit, offset int) domain.Page {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	if offset < 0 {
		offset = 0
	}

	return domain.Page{
		Limit:  limit,
		Offset: offset,
	}
}

// ============================================================================
// Response Builders
// ============================================================================

// BuildListEntriesResponse creates a ListEntriesResponse from a paged result.
func BuildListEntriesResponse(result *domain.PagedResult) *ListEntriesResponse {
	return &ListEntriesResponse{
		Entries:    EntriesToSummaryDTOs(result.Entries),
		TotalCount: result.Total,
		Limit:      result.Limit,
		Offset:     result.Offset,
		HasMore:    result.HasMore(),
	}
}

// BuildResourceTrailResponse creates a GetResourceTrailResponse from a paged result.
func BuildResourceTrailResponse(resourceType, resourceID string, result *domain.PagedResult) *GetResourceTrailResponse {
	return &GetResourceTrailResponse{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Entries:      EntriesToSummaryDTOs(result.Entries),
		TotalCount:   result.Total,
		Limit:        result.Limit,
		Offset:       result.Offset,
		HasMore:      result.HasMore(),
	}
}

// BuildActorActivityResponse creates a GetActorActivityResponse from a paged result.
func BuildActorActivityResponse(userID string, result *domain.PagedResult) *GetActorActivityResponse {
	return &GetActorActivityResponse{
		UserID:     userID,
		Entries:    EntriesToSummaryDTOs(result.Entries),
		TotalCount: result.Total,
		Limit:      result.Limit,
		Offset:     result.Offset,
		HasMore:    result.HasMore(),
	}
}
