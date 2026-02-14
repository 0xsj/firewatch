package detection

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/crypto"
	"github.com/0xsj/firewatch/pkg/timeutil"
)

// CampaignDetector groups related events into campaigns based
// on shared signatures, timing, and behavioral patterns.
type CampaignDetector struct {
	logger *slog.Logger
}

// NewCampaignDetector creates a campaign detector.
func NewCampaignDetector(logger *slog.Logger) *CampaignDetector {
	return &CampaignDetector{
		logger: logger.With("component", "campaign-detector"),
	}
}

// DetectCampaigns analyzes events and returns detected campaigns.
// Events should be pre-sorted by timestamp ascending.
func (cd *CampaignDetector) DetectCampaigns(events []*models.Event) []*models.Campaign {
	matches := cd.DetectCampaignsWithEvents(events)
	campaigns := make([]*models.Campaign, len(matches))
	for i, m := range matches {
		campaigns[i] = m.Campaign
	}
	return campaigns
}

// CampaignMatch pairs a detected campaign with the event IDs that belong to it.
type CampaignMatch struct {
	Campaign *models.Campaign
	EventIDs []string
}

// DetectCampaignsWithEvents analyzes events and returns detected campaigns
// along with the IDs of the events that belong to each campaign.
func (cd *CampaignDetector) DetectCampaignsWithEvents(events []*models.Event) []*CampaignMatch {
	if len(events) == 0 {
		return nil
	}

	// Strategy 1: Group by shared signature sets
	sigMatches := cd.groupBySignaturesWithEvents(events)

	// Strategy 2: Group by coordinated multi-IP attacks on same modules
	coordMatches := cd.groupByCoordinationWithEvents(events)

	matches := sigMatches
	matches = append(matches, coordMatches...)

	cd.logger.Info("campaign detection complete",
		"events", len(events),
		"campaigns", len(matches),
	)

	return matches
}

// groupBySignaturesWithEvents clusters events that share the same signature
// combination from multiple IPs, which suggests automated scanning.
func (cd *CampaignDetector) groupBySignaturesWithEvents(events []*models.Event) []*CampaignMatch {
	// Key: sorted signature set → events
	type cluster struct {
		events []*models.Event
		ips    map[string]struct{}
	}
	clusters := make(map[string]*cluster)

	for _, event := range events {
		if len(event.Signatures) == 0 {
			continue
		}
		key := signatureKey(event.Signatures)

		c, ok := clusters[key]
		if !ok {
			c = &cluster{ips: make(map[string]struct{})}
			clusters[key] = c
		}
		c.events = append(c.events, event)
		c.ips[event.SourceIP] = struct{}{}
	}

	matches := make([]*CampaignMatch, 0, len(clusters))
	for sigKey, c := range clusters {
		// Only flag as campaign if multiple IPs share the same sigs
		if len(c.ips) < 2 {
			continue
		}

		ips := setToSlice(c.ips)
		modules := uniqueModules(c.events)
		first, last := timeRange(c.events)

		eventIDs := make([]string, len(c.events))
		for i, e := range c.events {
			eventIDs[i] = e.ID
		}

		matches = append(matches, &CampaignMatch{
			Campaign: &models.Campaign{
				ID:              crypto.UUID4(),
				Name:            fmt.Sprintf("sig-cluster:%s", truncate(sigKey, 40)),
				FirstSeen:       first,
				LastSeen:        last,
				AttackerIPs:     ips,
				AttackerCount:   len(ips),
				EventCount:      len(c.events),
				ModulesTargeted: modules,
				Pattern:         fmt.Sprintf("shared signatures: %s", sigKey),
				Severity:        highestEventSeverity(c.events),
				Tags:            []string{"automated", "signature-cluster"},
			},
			EventIDs: eventIDs,
		})
	}

	return matches
}

// groupByCoordinationWithEvents detects multiple IPs hitting the same set
// of modules in a similar timeframe, suggesting coordinated scanning.
func (cd *CampaignDetector) groupByCoordinationWithEvents(events []*models.Event) []*CampaignMatch {
	// Key: IP → set of modules targeted
	ipModules := make(map[string]map[string]struct{})
	ipEvents := make(map[string][]*models.Event)

	for _, event := range events {
		if _, ok := ipModules[event.SourceIP]; !ok {
			ipModules[event.SourceIP] = make(map[string]struct{})
		}
		ipModules[event.SourceIP][event.Module] = struct{}{}
		ipEvents[event.SourceIP] = append(ipEvents[event.SourceIP], event)
	}

	// Group IPs that target the same module set
	type coordGroup struct {
		ips    []string
		events []*models.Event
	}
	groups := make(map[string]*coordGroup)

	for ip, modules := range ipModules {
		if len(modules) < 2 {
			continue // Single-module probes aren't interesting
		}
		key := moduleSetKey(modules)
		g, ok := groups[key]
		if !ok {
			g = &coordGroup{}
			groups[key] = g
		}
		g.ips = append(g.ips, ip)
		g.events = append(g.events, ipEvents[ip]...)
	}

	matches := make([]*CampaignMatch, 0, len(groups))
	for moduleKey, g := range groups {
		if len(g.ips) < 2 {
			continue
		}

		first, last := timeRange(g.events)
		modules := strings.Split(moduleKey, ",")

		eventIDs := make([]string, len(g.events))
		for i, e := range g.events {
			eventIDs[i] = e.ID
		}

		matches = append(matches, &CampaignMatch{
			Campaign: &models.Campaign{
				ID:              crypto.UUID4(),
				Name:            fmt.Sprintf("coordinated:%s", truncate(moduleKey, 40)),
				FirstSeen:       first,
				LastSeen:        last,
				AttackerIPs:     g.ips,
				AttackerCount:   len(g.ips),
				EventCount:      len(g.events),
				ModulesTargeted: modules,
				Pattern:         fmt.Sprintf("coordinated scanning of: %s", moduleKey),
				Severity:        highestEventSeverity(g.events),
				Tags:            []string{"coordinated", "multi-module"},
			},
			EventIDs: eventIDs,
		})
	}

	return matches
}

// signatureKey produces a stable key from a signature set.
func signatureKey(sigs []string) string {
	sorted := make([]string, len(sigs))
	copy(sorted, sigs)
	sort.Strings(sorted)
	return strings.Join(sorted, "+")
}

// moduleSetKey produces a stable key from a module set.
func moduleSetKey(modules map[string]struct{}) string {
	sorted := setToSlice(modules)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

// setToSlice converts a string set to a sorted slice.
func setToSlice(s map[string]struct{}) []string {
	result := make([]string, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// uniqueModules extracts the distinct module names from events.
func uniqueModules(events []*models.Event) []string {
	seen := make(map[string]struct{})
	for _, e := range events {
		seen[e.Module] = struct{}{}
	}
	return setToSlice(seen)
}

// timeRange returns the earliest and latest timestamps from events.
func timeRange(events []*models.Event) (first, last string) {
	if len(events) == 0 {
		now := timeutil.FormatRFC3339(timeutil.NowUTC())
		return now, now
	}
	first = events[0].Timestamp
	last = events[0].Timestamp
	for _, e := range events[1:] {
		if e.Timestamp < first {
			first = e.Timestamp
		}
		if e.Timestamp > last {
			last = e.Timestamp
		}
	}
	return first, last
}

// highestEventSeverity returns the highest severity across events.
func highestEventSeverity(events []*models.Event) string {
	highest := "info"
	for _, e := range events {
		if severityRank[e.Severity] > severityRank[highest] {
			highest = e.Severity
		}
	}
	return highest
}

// truncate shortens a string with ellipsis if it exceeds maxLen length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
