package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/0xsj/firewatch/internal/storage"
)

func runEvents(args []string) {
	fs := flag.NewFlagSet("events", flag.ExitOnError)
	configPath := fs.String("config", "firewatch.yaml", "path to config file")
	module := fs.String("module", "", "filter by module name")
	ip := fs.String("ip", "", "filter by source IP")
	severity := fs.String("severity", "", "filter by minimum severity")
	since := fs.String("since", "", "time filter (e.g., 1h, 24h, 7d, or RFC3339)")
	limit := fs.Int("limit", 50, "max results")
	asJSON := fs.Bool("json", false, "output as JSON")
	_ = fs.Parse(args)

	_, store, err := openStorage(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	filter := storage.EventFilter{
		SourceIP: *ip,
		Module:   *module,
		Severity: *severity,
		Limit:    *limit,
	}

	if *since != "" {
		t, err := parseSince(*since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --since value: %v\n", err)
			os.Exit(1)
		}
		filter.Since = t
	}

	events, err := store.ListEvents(context.Background(), filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing events: %v\n", err)
		os.Exit(1)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(events)
		return
	}

	if len(events) == 0 {
		fmt.Println("No events found.")
		return
	}

	// Table header.
	fmt.Printf("%-20s  %-15s  %-12s  %-8s  %-30s  %s\n",
		"TIMESTAMP", "SOURCE IP", "MODULE", "SEVERITY", "PATH", "SIGNATURES")
	fmt.Println(strings.Repeat("-", 110))

	for _, e := range events {
		ts := e.Timestamp
		if len(ts) > 19 {
			ts = ts[:19]
		}
		sigs := strings.Join(e.Signatures, ", ")
		if len(sigs) > 40 {
			sigs = sigs[:37] + "..."
		}
		path := e.Path
		if len(path) > 30 {
			path = path[:27] + "..."
		}
		fmt.Printf("%-20s  %-15s  %-12s  %-8s  %-30s  %s\n",
			ts, e.SourceIP, e.Module, e.Severity, path, sigs)
	}

	fmt.Printf("\nTotal: %d events\n", len(events))
}

// parseSince parses a duration string (e.g., "1h", "24h", "7d") or
// an RFC3339 timestamp into a time.Time.
func parseSince(s string) (time.Time, error) {
	// Try duration-like strings with "d" suffix.
	if strings.HasSuffix(s, "d") {
		days := s[:len(s)-1]
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err == nil {
			return time.Now().UTC().Add(-time.Duration(n) * 24 * time.Hour), nil
		}
	}

	// Try standard Go duration.
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().UTC().Add(-d), nil
	}

	// Try RFC3339.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try date-only.
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unrecognized time format %q (use 1h, 24h, 7d, or RFC3339)", s)
}
