package oauth

import (
	"context"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Repository defines the port for OAuth account persistence.
type Repository interface {
	// Save persists an OAuth account
	Save(ctx context.Context, account *Account) error

	// FindByID retrieves an OAuth account by its ID
	FindByID(ctx context.Context, id types.ID) (*Account, error)

	// FindByUserAndProvider finds an OAuth account by user and provider
	FindByUserAndProvider(ctx context.Context, userID types.ID, provider Provider) (*Account, error)

	// FindByProviderUserID finds an OAuth account by provider and provider's user ID
	FindByProviderUserID(ctx context.Context, provider Provider, providerUserID string) (*Account, error)

	// FindByUser retrieves all OAuth accounts for a user
	FindByUser(ctx context.Context, userID types.ID) ([]*Account, error)

	// Delete removes an OAuth account
	Delete(ctx context.Context, id types.ID) error
}
