package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetUserQuery handles retrieving a single user by ID.
// This is a read-only operation - no state changes.
type GetUserQuery struct {
	repo user.Repository
}

// NewGetUserQuery creates a new get user query.
func NewGetUserQuery(repo user.Repository) *GetUserQuery {
	return &GetUserQuery{
		repo: repo,
	}
}

// Handle executes the get user query.
func (q *GetUserQuery) Handle(ctx context.Context, userID types.ID) (*dto.UserDTO, error) {
	const op = "GetUserQuery.Handle"

	// 1. Validate input
	if userID.IsEmpty() {
		return nil, fmt.Errorf("%s: user ID is required", op)
	}

	// 2. Find user by ID
	// Repository automatically filters by tenant_id from context
	u, err := q.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// 3. Map to DTO
	return dto.MapUserToDTO(u), nil
}
