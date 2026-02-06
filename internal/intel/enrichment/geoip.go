package enrichment

import (
	"context"

	"github.com/0xsj/firewatch/internal/storage/models"
)

// GeoIP enriches IP-type IOCs with geolocation data.
// This is a placeholder — a real implementation would use
// MaxMind GeoLite2 or a similar database.
type GeoIP struct{}

func NewGeoIP() *GeoIP { return &GeoIP{} }

func (g *GeoIP) Name() string { return "geoip" }

func (g *GeoIP) Enrich(_ context.Context, ioc *models.IOC) error {
	if ioc.Type != models.IOCTypeIP {
		return nil
	}

	// Placeholder: a real implementation would look up the IP
	// in a GeoIP database (MaxMind, IP2Location, etc.).
	// For now, we leave GeoIP nil — callers should check for nil.
	//
	// Example with MaxMind:
	//   record, err := db.City(net.ParseIP(ioc.Value))
	//   ioc.GeoIP = &models.GeoIPInfo{
	//       Country:     record.Country.Names["en"],
	//       CountryCode: record.Country.IsoCode,
	//       City:        record.City.Names["en"],
	//       ASN:         record.Traits.AutonomousSystemNumber,
	//       ...
	//   }

	return nil
}
