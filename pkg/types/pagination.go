package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Default pagination limits
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// ============================================================================
// Cursor-Based Pagination (Recommended)
// ============================================================================

// Cursor represents an opaque pagination cursor.
// Cursors are base64-encoded strings that point to a specific position in a dataset.
//
// Example:
//
//	cursor := types.NewCursor("01HX...")  // Encode ID as cursor
//	id := cursor.Decode()                  // Decode back to ID
type Cursor string

// NewCursor creates a cursor from a string value (typically an ID).
// The value is base64-encoded to make it opaque.
func NewCursor(value string) Cursor {
	if value == "" {
		return ""
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(value))
	return Cursor(encoded)
}

// Decode decodes the cursor back to its original value.
// Returns empty string if cursor is invalid or empty.
func (c Cursor) Decode() string {
	if c == "" {
		return ""
	}

	decoded, err := base64.URLEncoding.DecodeString(string(c))
	if err != nil {
		return ""
	}

	return string(decoded)
}

// String returns the cursor as a string.
func (c Cursor) String() string {
	return string(c)
}

// IsEmpty returns true if the cursor is empty.
func (c Cursor) IsEmpty() bool {
	return c == ""
}

// IsValid returns true if the cursor can be decoded.
func (c Cursor) IsValid() bool {
	return c.Decode() != ""
}

// MarshalJSON implements json.Marshaler.
func (c Cursor) MarshalJSON() ([]byte, error) {
	if c.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(string(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Cursor) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*c = Cursor(s)
	return nil
}

// ============================================================================
// Cursor Pagination Request
// ============================================================================

// CursorPageRequest represents a cursor-based pagination request.
// Used for "infinite scroll" or "load more" patterns.
//
// Example:
//
//	req := types.CursorPageRequest{
//	    After: cursor,
//	    Limit: 20,
//	}
type CursorPageRequest struct {
	// After is the cursor to fetch records after (next page).
	// Empty means start from the beginning.
	After Cursor `json:"after,omitempty"`

	// Before is the cursor to fetch records before (previous page).
	// Empty means not fetching previous page.
	Before Cursor `json:"before,omitempty"`

	// Limit is the maximum number of records to return.
	// Defaults to DefaultPageSize if not specified.
	Limit int `json:"limit,omitempty"`
}

// Validate validates the pagination request and applies defaults.
func (r *CursorPageRequest) Validate() error {
	// Can't use both After and Before
	if !r.After.IsEmpty() && !r.Before.IsEmpty() {
		return fmt.Errorf("cannot specify both 'after' and 'before' cursors")
	}

	// Apply default limit
	if r.Limit <= 0 {
		r.Limit = DefaultPageSize
	}

	// Enforce max limit
	if r.Limit > MaxPageSize {
		r.Limit = MaxPageSize
	}

	// Enforce min limit
	if r.Limit < MinPageSize {
		r.Limit = MinPageSize
	}

	return nil
}

// Direction returns the pagination direction.
func (r CursorPageRequest) Direction() string {
	if !r.Before.IsEmpty() {
		return "backward"
	}
	return "forward"
}

// ============================================================================
// Cursor Pagination Response
// ============================================================================

// CursorPageResponse represents a cursor-based pagination response.
//
// Example:
//
//	response := types.CursorPageResponse[User]{
//	    Data:    users,
//	    NextCursor: types.NewCursor(lastUser.ID.String()),
//	    HasNext: len(users) == limit,
//	}
type CursorPageResponse[T any] struct {
	// Data is the list of records for this page.
	Data []T `json:"data"`

	// NextCursor is the cursor to fetch the next page.
	// Empty if there are no more records.
	NextCursor Cursor `json:"next_cursor,omitempty"`

	// PrevCursor is the cursor to fetch the previous page.
	// Empty if this is the first page.
	PrevCursor Cursor `json:"prev_cursor,omitempty"`

	// HasNext indicates if there are more records after this page.
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there are records before this page.
	HasPrev bool `json:"has_prev"`

	// Count is the number of records in this page.
	Count int `json:"count"`
}

// NewCursorPage creates a new cursor page response.
func NewCursorPage[T any](data []T, nextCursor, prevCursor Cursor, hasNext, hasPrev bool) CursorPageResponse[T] {
	return CursorPageResponse[T]{
		Data:       data,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Count:      len(data),
	}
}

// IsEmpty returns true if the page has no data.
func (p CursorPageResponse[T]) IsEmpty() bool {
	return len(p.Data) == 0
}

// ============================================================================
// Offset-Based Pagination (Simple but less efficient)
// ============================================================================

// OffsetPageRequest represents an offset-based pagination request.
// Simpler than cursor-based but less efficient for large datasets.
//
// Example:
//
//	req := types.OffsetPageRequest{
//	    Page:     1,
//	    PageSize: 20,
//	}
type OffsetPageRequest struct {
	// Page is the page number (1-indexed).
	// Defaults to 1 if not specified.
	Page int `json:"page,omitempty"`

	// PageSize is the number of records per page.
	// Defaults to DefaultPageSize if not specified.
	PageSize int `json:"page_size,omitempty"`
}

// Validate validates the pagination request and applies defaults.
func (r *OffsetPageRequest) Validate() error {
	// Apply default page
	if r.Page <= 0 {
		r.Page = 1
	}

	// Apply default page size
	if r.PageSize <= 0 {
		r.PageSize = DefaultPageSize
	}

	// Enforce max page size
	if r.PageSize > MaxPageSize {
		r.PageSize = MaxPageSize
	}

	// Enforce min page size
	if r.PageSize < MinPageSize {
		r.PageSize = MinPageSize
	}

	return nil
}

// Offset returns the offset for database queries.
// Offset = (Page - 1) * PageSize
func (r OffsetPageRequest) Offset() int {
	return (r.Page - 1) * r.PageSize
}

// Limit returns the limit for database queries.
func (r OffsetPageRequest) Limit() int {
	return r.PageSize
}

// ============================================================================
// Offset Pagination Response
// ============================================================================

// OffsetPageResponse represents an offset-based pagination response.
//
// Example:
//
//	response := types.OffsetPageResponse[User]{
//	    Data:       users,
//	    Page:       1,
//	    PageSize:   20,
//	    TotalCount: 150,
//	}
type OffsetPageResponse[T any] struct {
	// Data is the list of records for this page.
	Data []T `json:"data"`

	// Page is the current page number (1-indexed).
	Page int `json:"page"`

	// PageSize is the number of records per page.
	PageSize int `json:"page_size"`

	// TotalCount is the total number of records across all pages.
	TotalCount int64 `json:"total_count"`

	// TotalPages is the total number of pages.
	TotalPages int `json:"total_pages"`

	// HasNext indicates if there is a next page.
	HasNext bool `json:"has_next"`

	// HasPrev indicates if there is a previous page.
	HasPrev bool `json:"has_prev"`
}

// NewOffsetPage creates a new offset page response.
func NewOffsetPage[T any](data []T, page, pageSize int, totalCount int64) OffsetPageResponse[T] {
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 0 {
		totalPages = 0
	}

	return OffsetPageResponse[T]{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// IsEmpty returns true if the page has no data.
func (p OffsetPageResponse[T]) IsEmpty() bool {
	return len(p.Data) == 0
}

// IsFirstPage returns true if this is the first page.
func (p OffsetPageResponse[T]) IsFirstPage() bool {
	return p.Page == 1
}

// IsLastPage returns true if this is the last page.
func (p OffsetPageResponse[T]) IsLastPage() bool {
	return p.Page >= p.TotalPages
}

// NextPage returns the next page number.
// Returns 0 if there is no next page.
func (p OffsetPageResponse[T]) NextPage() int {
	if !p.HasNext {
		return 0
	}
	return p.Page + 1
}

// PrevPage returns the previous page number.
// Returns 0 if there is no previous page.
func (p OffsetPageResponse[T]) PrevPage() int {
	if !p.HasPrev {
		return 0
	}
	return p.Page - 1
}

// ============================================================================
// Helper Functions
// ============================================================================

// ParseCursorPageRequest parses cursor pagination from query parameters.
// Expects: ?after=xxx&limit=20 or ?before=xxx&limit=20
func ParseCursorPageRequest(after, before string, limit string) (CursorPageRequest, error) {
	req := CursorPageRequest{
		After:  Cursor(after),
		Before: Cursor(before),
	}

	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return req, fmt.Errorf("invalid limit: %w", err)
		}
		req.Limit = l
	}

	if err := req.Validate(); err != nil {
		return req, err
	}

	return req, nil
}

// ParseOffsetPageRequest parses offset pagination from query parameters.
// Expects: ?page=1&page_size=20
func ParseOffsetPageRequest(page, pageSize string) (OffsetPageRequest, error) {
	req := OffsetPageRequest{}

	if page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			return req, fmt.Errorf("invalid page: %w", err)
		}
		req.Page = p
	}

	if pageSize != "" {
		ps, err := strconv.Atoi(pageSize)
		if err != nil {
			return req, fmt.Errorf("invalid page_size: %w", err)
		}
		req.PageSize = ps
	}

	if err := req.Validate(); err != nil {
		return req, err
	}

	return req, nil
}

// ============================================================================
// Sorting
// ============================================================================

// SortOrder represents the sort direction.
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// String returns the string representation.
func (s SortOrder) String() string {
	return string(s)
}

// IsValid returns true if the sort order is valid.
func (s SortOrder) IsValid() bool {
	return s == SortAsc || s == SortDesc
}

// Reverse returns the opposite sort order.
func (s SortOrder) Reverse() SortOrder {
	if s == SortAsc {
		return SortDesc
	}
	return SortAsc
}

// ParseSortOrder parses a sort order string.
// Returns SortAsc if the string is invalid.
func ParseSortOrder(s string) SortOrder {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "asc", "ascending":
		return SortAsc
	case "desc", "descending":
		return SortDesc
	default:
		return SortAsc // Default to ascending
	}
}

// SortField represents a field to sort by with its order.
type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

// NewSortField creates a new sort field.
func NewSortField(field string, order SortOrder) SortField {
	return SortField{
		Field: field,
		Order: order,
	}
}

// String returns a string representation (e.g., "created_at:desc").
func (s SortField) String() string {
	return fmt.Sprintf("%s:%s", s.Field, s.Order)
}

// ParseSortField parses a sort field string.
// Format: "field:order" (e.g., "created_at:desc")
// If no order is specified, defaults to ascending.
func ParseSortField(s string) SortField {
	parts := strings.Split(s, ":")
	field := strings.TrimSpace(parts[0])

	order := SortAsc
	if len(parts) > 1 {
		order = ParseSortOrder(parts[1])
	}

	return SortField{
		Field: field,
		Order: order,
	}
}

// SortFields represents multiple sort fields.
type SortFields []SortField

// ParseSortFields parses multiple sort fields from a comma-separated string.
// Format: "field1:order1,field2:order2"
// Example: "created_at:desc,name:asc"
func ParseSortFields(s string) SortFields {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	fields := make(SortFields, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			fields = append(fields, ParseSortField(part))
		}
	}

	return fields
}

// String returns a comma-separated string representation.
func (sf SortFields) String() string {
	if len(sf) == 0 {
		return ""
	}

	parts := make([]string, len(sf))
	for i, field := range sf {
		parts[i] = field.String()
	}

	return strings.Join(parts, ",")
}
