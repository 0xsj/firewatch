package export

import (
	"encoding/json"
	"fmt"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// STIX exports threat intelligence in STIX 2.1 format.
// See: https://docs.oasis-open.org/cti/stix/v2.1/stix-v2.1.html
type STIX struct{}

func NewSTIX() *STIX { return &STIX{} }

func (s *STIX) Name() string        { return "stix" }
func (s *STIX) ContentType() string { return "application/json" }

// stixBundle is a STIX 2.1 Bundle containing objects.
type stixBundle struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Objects []any  `json:"objects"`
}

// stixIndicator represents a STIX 2.1 Indicator SDO.
type stixIndicator struct {
	Type        string   `json:"type"`
	SpecVersion string   `json:"spec_version"`
	ID          string   `json:"id"`
	Created     string   `json:"created"`
	Modified    string   `json:"modified"`
	Name        string   `json:"name"`
	Pattern     string   `json:"pattern"`
	PatternType string   `json:"pattern_type"`
	ValidFrom   string   `json:"valid_from"`
	Labels      []string `json:"labels,omitempty"`
}

// stixCampaign represents a STIX 2.1 Campaign SDO.
type stixCampaign struct {
	Type        string `json:"type"`
	SpecVersion string `json:"spec_version"`
	ID          string `json:"id"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FirstSeen   string `json:"first_seen"`
	LastSeen    string `json:"last_seen"`
}

func (s *STIX) ExportIOCs(iocs []*models.IOC) ([]byte, error) {
	bundle := stixBundle{
		Type:    "bundle",
		ID:      fmt.Sprintf("bundle--%s", crypto.UUID4()),
		Objects: make([]any, 0, len(iocs)),
	}

	now := timeutil.FormatRFC3339(timeutil.NowUTC())

	for _, ioc := range iocs {
		indicator := stixIndicator{
			Type:        "indicator",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("indicator--%s", ioc.ID),
			Created:     now,
			Modified:    now,
			Name:        fmt.Sprintf("%s: %s", ioc.Type, ioc.Value),
			Pattern:     iocToSTIXPattern(ioc),
			PatternType: "stix",
			ValidFrom:   ioc.FirstSeen,
			Labels:      ioc.Tags,
		}
		bundle.Objects = append(bundle.Objects, indicator)
	}

	return json.MarshalIndent(bundle, "", "  ")
}

func (s *STIX) ExportCampaigns(campaigns []*models.Campaign) ([]byte, error) {
	bundle := stixBundle{
		Type:    "bundle",
		ID:      fmt.Sprintf("bundle--%s", crypto.UUID4()),
		Objects: make([]any, 0, len(campaigns)),
	}

	now := timeutil.FormatRFC3339(timeutil.NowUTC())

	for _, campaign := range campaigns {
		obj := stixCampaign{
			Type:        "campaign",
			SpecVersion: "2.1",
			ID:          fmt.Sprintf("campaign--%s", campaign.ID),
			Created:     now,
			Modified:    now,
			Name:        campaign.Name,
			Description: campaign.Pattern,
			FirstSeen:   campaign.FirstSeen,
			LastSeen:    campaign.LastSeen,
		}
		bundle.Objects = append(bundle.Objects, obj)
	}

	return json.MarshalIndent(bundle, "", "  ")
}

// iocToSTIXPattern converts an IOC to a STIX 2.1 pattern string.
func iocToSTIXPattern(ioc *models.IOC) string {
	switch ioc.Type {
	case models.IOCTypeIP:
		return fmt.Sprintf("[ipv4-addr:value = '%s']", ioc.Value)
	case models.IOCTypeDomain:
		return fmt.Sprintf("[domain-name:value = '%s']", ioc.Value)
	case models.IOCTypeURL:
		return fmt.Sprintf("[url:value = '%s']", ioc.Value)
	case models.IOCTypeHash:
		return fmt.Sprintf("[file:hashes.SHA-256 = '%s']", ioc.Value)
	case models.IOCTypeEmail:
		return fmt.Sprintf("[email-addr:value = '%s']", ioc.Value)
	case models.IOCTypeCIDR:
		return fmt.Sprintf("[ipv4-addr:value = '%s']", ioc.Value)
	default:
		return fmt.Sprintf("[artifact:payload_bin = '%s']", ioc.Value)
	}
}
