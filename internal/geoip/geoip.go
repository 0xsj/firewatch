package geoip

import (
	"net"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/oschwald/geoip2-golang"
)

// Reader wraps a MaxMind GeoIP2 database for IP geolocation lookups.
type Reader struct {
	db *geoip2.Reader
}

// Open opens a MaxMind .mmdb file (e.g., GeoLite2-City.mmdb).
func Open(path string) (*Reader, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &Reader{db: db}, nil
}

// Lookup returns geolocation info for the given IP address.
// Returns (nil, nil) for private, loopback, or invalid IPs.
func (r *Reader) Lookup(ipStr string) (*models.GeoIPInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, nil
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
		return nil, nil
	}

	record, err := r.db.City(ip)
	if err != nil {
		return nil, err
	}

	return &models.GeoIPInfo{
		Country:     record.Country.Names["en"],
		CountryCode: record.Country.IsoCode,
		City:        record.City.Names["en"],
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		// ASN fields require a separate GeoLite2-ASN.mmdb database.
	}, nil
}

// Close releases the underlying database resources.
func (r *Reader) Close() error {
	return r.db.Close()
}
