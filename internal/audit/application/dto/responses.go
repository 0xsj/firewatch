// internal/audit/application/dto/responses.go
package dto

// GetEntryResponse represents the output of getting a single audit entry.
type GetEntryResponse struct {
	Entry AuditEntryDTO `json:"entry"`
}

// ListEntriesResponse represents a paginated list of audit entries.
type ListEntriesResponse struct {
	Entries    []AuditEntrySummaryDTO `json:"entries"`
	TotalCount int                    `json:"total_count"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	HasMore    bool                   `json:"has_more"`
}

// GetResourceTrailResponse represents the output of getting a resource's audit trail.
type GetResourceTrailResponse struct {
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Entries      []AuditEntrySummaryDTO `json:"entries"`
	TotalCount   int                    `json:"total_count"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
	HasMore      bool                   `json:"has_more"`
}

// GetActorActivityResponse represents the output of getting an actor's activity.
type GetActorActivityResponse struct {
	UserID     string                 `json:"user_id"`
	Entries    []AuditEntrySummaryDTO `json:"entries"`
	TotalCount int                    `json:"total_count"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	HasMore    bool                   `json:"has_more"`
}
