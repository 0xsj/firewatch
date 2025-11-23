package dto

import (
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
)

// MapUserToDTO maps a domain user to a DTO.
// Converts domain types to primitives suitable for API responses.
func MapUserToDTO(u *user.User) *UserDTO {
	if u == nil {
		return nil
	}

	dto := &UserDTO{
		ID:            u.ID().String(),
		TenantID:      u.TenantID(),
		Email:         u.Email().String(),
		Status:        u.Status().String(),
		Role:          u.Role().String(),
		EmailVerified: u.EmailVerified(),
		CreatedAt:     u.CreatedAt().Time(),
		UpdatedAt:     u.UpdatedAt().Time(),
	}

	// Map optional fields (only if not nil)
	if u.EmailVerifiedAt() != nil {
		t := u.EmailVerifiedAt().Time()
		dto.EmailVerifiedAt = &t
	}

	if u.LastLoginAt() != nil {
		t := u.LastLoginAt().Time()
		dto.LastLoginAt = &t
	}

	return dto
}

// MapUserToSummaryDTO maps a domain user to a summary DTO.
// Contains only essential fields for list views.
func MapUserToSummaryDTO(u *user.User) *UserSummaryDTO {
	if u == nil {
		return nil
	}

	return &UserSummaryDTO{
		ID:        u.ID().String(),
		Email:     u.Email().String(),
		Status:    u.Status().String(),
		Role:      u.Role().String(),
		CreatedAt: u.CreatedAt().Time(),
	}
}

// MapUsersToSummaryDTOs maps multiple domain users to summary DTOs.
// Helper for list operations.
func MapUsersToSummaryDTOs(users []*user.User) []*UserSummaryDTO {
	if users == nil {
		return nil
	}

	dtos := make([]*UserSummaryDTO, len(users))
	for i, u := range users {
		dtos[i] = MapUserToSummaryDTO(u)
	}
	return dtos
}

// MapFiltersFromRequest maps a list request to domain filters.
// Converts DTO types to domain filter types.
func MapFiltersFromRequest(req ListUsersRequest) user.Filters {
	filters := user.DefaultFilters()

	// Apply status filter if provided
	if req.Status != nil {
		if status, err := user.ParseStatus(*req.Status); err == nil {
			filters = filters.WithStatus(status)
		}
	}

	// Apply role filter if provided
	if req.Role != nil {
		if role, err := user.ParseRole(*req.Role); err == nil {
			filters = filters.WithRole(role)
		}
	}

	// Apply email verified filter if provided
	if req.EmailVerified != nil {
		filters = filters.WithEmailVerified(*req.EmailVerified)
	}

	// Apply email search filter
	if req.EmailContains != "" {
		filters = filters.WithEmailContains(req.EmailContains)
	}

	// Apply pagination
	limit := req.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}
	filters = filters.WithPagination(limit, req.Offset)

	// Apply sorting
	if req.SortBy != "" {
		sortBy := mapSortField(req.SortBy)
		sortOrder := mapSortOrder(req.SortOrder)
		filters = filters.WithSort(sortBy, sortOrder)
	}

	return filters
}

// MapListUsersResponse creates a paginated response from domain users.
func MapListUsersResponse(users []*user.User, totalCount int, req ListUsersRequest) *ListUsersResponse {
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}

	return &ListUsersResponse{
		Users:      MapUsersToSummaryDTOs(users),
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     req.Offset,
		HasMore:    req.Offset+len(users) < totalCount,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// mapSortField converts a string sort field to domain type.
func mapSortField(s string) user.SortField {
	switch s {
	case "created_at":
		return user.SortByCreatedAt
	case "updated_at":
		return user.SortByUpdatedAt
	case "email":
		return user.SortByEmail
	case "last_login_at":
		return user.SortByLastLoginAt
	default:
		return user.SortByCreatedAt // Default
	}
}

// mapSortOrder converts a string sort order to domain type.
func mapSortOrder(s string) user.SortOrder {
	switch s {
	case "asc":
		return user.SortOrderAsc
	case "desc":
		return user.SortOrderDesc
	default:
		return user.SortOrderDesc // Default
	}
}
