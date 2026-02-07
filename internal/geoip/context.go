package geoip

import (
	"context"

	"github.com/0xsj/firewatch/internal/storage/models"
)

type contextKey string

const geoIPKey contextKey = "geoip_info"

// WithGeoIP stores GeoIP lookup results in the context.
func WithGeoIP(ctx context.Context, info *models.GeoIPInfo) context.Context {
	return context.WithValue(ctx, geoIPKey, info)
}

// FromContext retrieves GeoIP info from the context.
// Returns nil if no GeoIP data is present.
func FromContext(ctx context.Context) *models.GeoIPInfo {
	info, _ := ctx.Value(geoIPKey).(*models.GeoIPInfo)
	return info
}
