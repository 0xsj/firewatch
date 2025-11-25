package query

import (
	"context"
	"fmt"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/session"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// ListSessionsQuery handles listing active sessions for a user.
type ListSessionsQuery struct {
	repo session.Repository
}

// NewListSessionsQuery creates a new ListSessionsQuery.
func NewListSessionsQuery(repo session.Repository) *ListSessionsQuery {
	return &ListSessionsQuery{
		repo: repo,
	}
}

// ListSessionsRequest is the input for listing sessions.
type ListSessionsRequest struct {
	UserID   types.ID // From JWT claims
	TenantID string   // From JWT claims (optional filtering)
}

// Handle executes the query to list active sessions.
func (q *ListSessionsQuery) Handle(ctx context.Context, req ListSessionsRequest) ([]*dto.SessionDTO, error) {
	const op = "ListSessionsQuery.Handle"

	var sessions []*session.Session
	var err error

	if req.TenantID != "" {
		sessions, err = q.repo.FindActiveByUserIDAndTenant(ctx, req.UserID, req.TenantID)
	} else {
		sessions, err = q.repo.FindActiveByUserID(ctx, req.UserID)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Convert to DTOs
	result := make([]*dto.SessionDTO, len(sessions))
	for i, s := range sessions {
		result[i] = dto.NewSessionResponse(s)
	}

	return result, nil
}
