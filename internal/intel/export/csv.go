package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/0xsj/firewatch/internal/storage/models"
)

// CSV exports threat intelligence as comma-separated values.
type CSV struct{}

func NewCSV() *CSV { return &CSV{} }

func (c *CSV) Name() string        { return "csv" }
func (c *CSV) ContentType() string { return "text/csv" }

func (c *CSV) ExportIOCs(iocs []*models.IOC) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	if err := w.Write([]string{
		"id", "type", "value", "severity", "first_seen", "last_seen",
		"hostname", "country", "asn", "tags",
	}); err != nil {
		return nil, fmt.Errorf("writing CSV header: %w", err)
	}

	for _, ioc := range iocs {
		country := ""
		asn := ""
		if ioc.GeoIP != nil {
			country = ioc.GeoIP.CountryCode
			asn = fmt.Sprintf("AS%d", ioc.GeoIP.ASN)
		}

		if err := w.Write([]string{
			ioc.ID,
			string(ioc.Type),
			ioc.Value,
			ioc.Severity,
			ioc.FirstSeen,
			ioc.LastSeen,
			ioc.Hostname,
			country,
			asn,
			strings.Join(ioc.Tags, ";"),
		}); err != nil {
			return nil, fmt.Errorf("writing CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flushing CSV: %w", err)
	}

	return buf.Bytes(), nil
}

func (c *CSV) ExportCampaigns(campaigns []*models.Campaign) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	if err := w.Write([]string{
		"id", "name", "severity", "first_seen", "last_seen",
		"attacker_count", "event_count", "modules_targeted", "pattern", "tags",
	}); err != nil {
		return nil, fmt.Errorf("writing CSV header: %w", err)
	}

	for _, campaign := range campaigns {
		if err := w.Write([]string{
			campaign.ID,
			campaign.Name,
			campaign.Severity,
			campaign.FirstSeen,
			campaign.LastSeen,
			fmt.Sprintf("%d", campaign.AttackerCount),
			fmt.Sprintf("%d", campaign.EventCount),
			strings.Join(campaign.ModulesTargeted, ";"),
			campaign.Pattern,
			strings.Join(campaign.Tags, ";"),
		}); err != nil {
			return nil, fmt.Errorf("writing CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flushing CSV: %w", err)
	}

	return buf.Bytes(), nil
}
