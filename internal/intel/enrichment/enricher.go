package enrichment

import (
	"context"

	"github.com/0xsj/firewatch/internal/storage/models"
)

// Enricher adds context to IOCs from external data sources.
type Enricher interface {
	// Name returns the enricher's identifier.
	Name() string

	// Enrich adds data to the IOC in-place. It should only modify
	// fields relevant to this enricher, leaving others untouched.
	// Returns an error if enrichment fails (callers may ignore it).
	Enrich(ctx context.Context, ioc *models.IOC) error
}
