package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/0xsj/firewatch/internal/intel"
	"github.com/0xsj/firewatch/internal/intel/enrichment"
	"github.com/0xsj/firewatch/internal/intel/export"
	"github.com/0xsj/firewatch/internal/storage"
)

func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	configPath := fs.String("config", "firewatch.yaml", "path to config file")
	format := fs.String("format", "csv", "export format: csv, stix, misp")
	since := fs.String("since", "", "time filter (e.g., 1h, 24h, 7d, or RFC3339)")
	output := fs.String("output", "", "output file (default: stdout)")
	fs.String("o", "", "alias for --output")
	_ = fs.Parse(args)

	// Handle -o alias.
	if *output == "" {
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "o" {
				*output = f.Value.String()
			}
		})
	}

	cfg, store, err := openStorage(*configPath)
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

	// Build enrichers.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	geoReader := openGeoIP(cfg, logger)
	if geoReader != nil {
		defer geoReader.Close()
	}

	enrichers := []enrichment.Enricher{
		enrichment.NewGeoIP(geoReader),
		enrichment.NewDNS(),
	}

	collector := intel.NewCollector(store, enrichers, logger)

	ctx := context.Background()
	result, err := collector.Collect(ctx, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error collecting intel: %v\n", err)
		os.Exit(1)
	}

	// Pick exporter.
	var exporter export.Exporter
	switch *format {
	case "csv":
		exporter = export.NewCSV()
	case "stix":
		exporter = export.NewSTIX()
	case "misp":
		exporter = export.NewMISP()
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q (use csv, stix, or misp)\n", *format)
		os.Exit(1)
	}

	// Export IOCs.
	iocData, err := exporter.ExportIOCs(result.IOCs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error exporting IOCs: %v\n", err)
		os.Exit(1)
	}

	// Export campaigns.
	campaignData, err := exporter.ExportCampaigns(result.Campaigns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error exporting campaigns: %v\n", err)
		os.Exit(1)
	}

	// Write output.
	var w io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	_, _ = w.Write(iocData)
	if len(result.Campaigns) > 0 {
		if *format == "csv" {
			_, _ = w.Write([]byte("\n"))
		}
		_, _ = w.Write(campaignData)
	}

	if *output != "" {
		fmt.Fprintf(os.Stderr, "exported %d IOCs and %d campaigns to %s\n",
			len(result.IOCs), len(result.Campaigns), *output)
	}
}
