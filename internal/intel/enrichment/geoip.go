package enrichment

import (
	"context"

	"github.com/0xsj/firewatch/internal/geoip"
	"github.com/0xsj/firewatch/internal/storage/models"
)

// GeoIP enriches IP-type IOCs with geolocation data using a
// MaxMind GeoIP2 database.
type GeoIP struct {
	reader *geoip.Reader
}

// NewGeoIP creates a GeoIP enricher. If reader is nil, Enrich
// is a no-op (graceful degradation when no DB is available).
func NewGeoIP(reader *geoip.Reader) *GeoIP {
	return &GeoIP{reader: reader}
}

func (g *GeoIP) Name() string { return "geoip" }

func (g *GeoIP) Enrich(_ context.Context, ioc *models.IOC) error {
	if g.reader == nil {
		return nil
	}
	if ioc.Type != models.IOCTypeIP {
		return nil
	}

	info, err := g.reader.Lookup(ioc.Value)
	if err != nil {
		return err
	}
	ioc.GeoIP = info
	return nil
}
