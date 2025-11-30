package provider

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/database"
	"github.com/0xsj/hexagonal-go/pkg/flags"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// ProvideFlagsClient creates a new feature flags client.
func ProvideFlagsClient(db database.DB, log logger.Logger) flags.Client {
	repo := flags.NewRepositoryAdapter(db)

	return flags.NewClient(
		repo,
		log,
		flags.WithCache(10*time.Second), // Short TTL for development
		flags.WithDefaultTenant("default"),
	)
}
