package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/0xsj/firewatch/internal/storage"
)

type statsResult struct {
	TotalEvents int64          `json:"total_events"`
	UniqueIPs   int            `json:"unique_ips"`
	ByModule    map[string]int `json:"by_module"`
	BySeverity  map[string]int `json:"by_severity"`
	TopIPs      []ipCount      `json:"top_ips"`
	TopSigs     []sigCount     `json:"top_signatures"`
}

type ipCount struct {
	IP    string `json:"ip"`
	Count int    `json:"count"`
}

type sigCount struct {
	Signature string `json:"signature"`
	Count     int    `json:"count"`
}

func runStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	configPath := fs.String("config", "firewatch.yaml", "path to config file")
	since := fs.String("since", "24h", "time window (e.g., 1h, 24h, 7d)")
	asJSON := fs.Bool("json", false, "output as JSON")
	_ = fs.Parse(args)

	_, store, err := openStorage(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	filter := storage.EventFilter{}
	if *since != "" {
		t, err := parseSince(*since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid --since value: %v\n", err)
			os.Exit(1)
		}
		filter.Since = t
	}

	ctx := context.Background()

	total, err := store.CountEvents(ctx, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error counting events: %v\n", err)
		os.Exit(1)
	}

	events, err := store.ListEvents(ctx, storage.EventFilter{Since: filter.Since})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing events: %v\n", err)
		os.Exit(1)
	}

	// Aggregate.
	byModule := make(map[string]int)
	bySeverity := make(map[string]int)
	byIP := make(map[string]int)
	bySig := make(map[string]int)

	for _, e := range events {
		byModule[e.Module]++
		bySeverity[e.Severity]++
		byIP[e.SourceIP]++
		for _, sig := range e.Signatures {
			bySig[sig]++
		}
	}

	topIPs := sortedCounts(byIP, 10)
	topSigs := sortedSigCounts(bySig, 10)

	result := statsResult{
		TotalEvents: total,
		UniqueIPs:   len(byIP),
		ByModule:    byModule,
		BySeverity:  bySeverity,
		TopIPs:      topIPs,
		TopSigs:     topSigs,
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return
	}

	// Pretty print.
	fmt.Printf("Firewatch Stats (last %s)\n", *since)
	fmt.Println(strings.Repeat("\u2500", 40))
	fmt.Printf("Events:     %d\n", total)
	fmt.Printf("Unique IPs: %d\n", len(byIP))

	if len(byModule) > 0 {
		fmt.Println("\nBy Module:")
		for _, kv := range sortedCounts(byModule, 0) {
			pct := 0
			if total > 0 {
				pct = int(float64(kv.Count) / float64(total) * 100)
			}
			fmt.Printf("  %-16s %5d (%d%%)\n", kv.IP, kv.Count, pct)
		}
	}

	if len(bySeverity) > 0 {
		fmt.Println("\nBy Severity:")
		for _, sev := range []string{"critical", "high", "medium", "low", "info"} {
			if c, ok := bySeverity[sev]; ok {
				fmt.Printf("  %-16s %5d\n", sev, c)
			}
		}
	}

	if len(topIPs) > 0 {
		fmt.Println("\nTop Source IPs:")
		for _, kv := range topIPs {
			fmt.Printf("  %-18s %5d events\n", kv.IP, kv.Count)
		}
	}

	if len(topSigs) > 0 {
		fmt.Println("\nTop Signatures:")
		for _, kv := range topSigs {
			fmt.Printf("  %-24s %5d hits\n", kv.Signature, kv.Count)
		}
	}
}

func sortedCounts(m map[string]int, limit int) []ipCount {
	result := make([]ipCount, 0, len(m))
	for k, v := range m {
		result = append(result, ipCount{IP: k, Count: v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func sortedSigCounts(m map[string]int, limit int) []sigCount {
	result := make([]sigCount, 0, len(m))
	for k, v := range m {
		result = append(result, sigCount{Signature: k, Count: v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}
