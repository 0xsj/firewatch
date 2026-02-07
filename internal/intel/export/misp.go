package export

import (
	"encoding/json"
	"fmt"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// MISP exports threat intelligence in MISP event format.
// See: https://www.misp-project.org/datamodels/
type MISP struct{}

func NewMISP() *MISP { return &MISP{} }

func (m *MISP) Name() string        { return "misp" }
func (m *MISP) ContentType() string { return "application/json" }

// mispEvent is a MISP event container.
type mispEvent struct {
	Info        string          `json:"info"`
	ThreatLevel string          `json:"threat_level_id"`
	Date        string          `json:"date"`
	Published   bool            `json:"published"`
	Attributes  []mispAttribute `json:"Attribute"`
}

// mispAttribute is a single MISP attribute (IOC).
type mispAttribute struct {
	UUID     string `json:"uuid"`
	Type     string `json:"type"`
	Category string `json:"category"`
	Value    string `json:"value"`
	ToIDs    bool   `json:"to_ids"`
	Comment  string `json:"comment,omitempty"`
}

func (m *MISP) ExportIOCs(iocs []*models.IOC) ([]byte, error) {
	attrs := make([]mispAttribute, 0, len(iocs))

	for _, ioc := range iocs {
		mispType, category := iocToMISPType(ioc)

		attrs = append(attrs, mispAttribute{
			UUID:     ioc.ID,
			Type:     mispType,
			Category: category,
			Value:    ioc.Value,
			ToIDs:    ioc.Severity == "high" || ioc.Severity == "critical",
			Comment:  fmt.Sprintf("Source: firewatch, Severity: %s", ioc.Severity),
		})
	}

	event := mispEvent{
		Info:        fmt.Sprintf("Firewatch honeypot IOCs - %s", timeutil.FormatRFC3339(timeutil.NowUTC())[:10]),
		ThreatLevel: mispThreatLevel(iocs),
		Date:        timeutil.FormatRFC3339(timeutil.NowUTC())[:10],
		Published:   false,
		Attributes:  attrs,
	}

	return json.MarshalIndent(event, "", "  ")
}

func (m *MISP) ExportCampaigns(campaigns []*models.Campaign) ([]byte, error) {
	events := make([]mispEvent, 0, len(campaigns))

	for _, campaign := range campaigns {
		attrs := make([]mispAttribute, 0, len(campaign.AttackerIPs))

		for _, ip := range campaign.AttackerIPs {
			attrs = append(attrs, mispAttribute{
				UUID:     crypto.UUID4(),
				Type:     "ip-src",
				Category: "Network activity",
				Value:    ip,
				ToIDs:    true,
				Comment:  fmt.Sprintf("Campaign: %s", campaign.Name),
			})
		}

		events = append(events, mispEvent{
			Info:        campaign.Name,
			ThreatLevel: severityToMISPThreat(campaign.Severity),
			Date:        campaign.FirstSeen[:10],
			Published:   false,
			Attributes:  attrs,
		})
	}

	return json.MarshalIndent(events, "", "  ")
}

// iocToMISPType maps internal IOC types to MISP attribute types.
func iocToMISPType(ioc *models.IOC) (mispType, category string) {
	switch ioc.Type {
	case models.IOCTypeIP:
		return "ip-src", "Network activity"
	case models.IOCTypeDomain:
		return "domain", "Network activity"
	case models.IOCTypeURL:
		return "url", "Network activity"
	case models.IOCTypeHash:
		return "sha256", "Payload delivery"
	case models.IOCTypeEmail:
		return "email-src", "Payload delivery"
	case models.IOCTypeCIDR:
		return "ip-src", "Network activity"
	default:
		return "text", "Other"
	}
}

// mispThreatLevel returns the highest threat level from a set of IOCs.
// MISP threat levels: 1=High, 2=Medium, 3=Low, 4=Undefined.
func mispThreatLevel(iocs []*models.IOC) string {
	highest := 0
	for _, ioc := range iocs {
		rank := severityRankMap[ioc.Severity]
		if rank > highest {
			highest = rank
		}
	}
	return severityToMISPThreat(rankToSeverity(highest))
}

func severityToMISPThreat(severity string) string {
	switch severity {
	case "critical", "high":
		return "1"
	case "medium":
		return "2"
	case "low":
		return "3"
	default:
		return "4"
	}
}

var severityRankMap = map[string]int{
	"info":     0,
	"low":      1,
	"medium":   2,
	"high":     3,
	"critical": 4,
}

func rankToSeverity(rank int) string {
	for sev, r := range severityRankMap {
		if r == rank {
			return sev
		}
	}
	return "info"
}
