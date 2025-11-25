package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// GetCurrentUserQuery handles fetching the currently authenticated user.
type GetCurrentUserQuery struct {
	repo user.Repository
}

// NewGetCurrentUserQuery creates a new GetCurrentUserQuery.
func NewGetCurrentUserQuery(repo user.Repository) *GetCurrentUserQuery {
	return &GetCurrentUserQuery{
		repo: repo,
	}
}

// GetCurrentUserRequest is the input for getting the current user.
type GetCurrentUserRequest struct {
	UserID types.ID // From JWT claims
}

// Handle executes the query to get the current user.
func (q *GetCurrentUserQuery) Handle(ctx context.Context, req GetCurrentUserRequest) (*dto.UserDTO, error) {
	const op = "GetCurrentUserQuery.Handle"

	u, err := q.repo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return dto.NewUserResponse(u), nil
}
