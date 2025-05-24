package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/aggregator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/coordinator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

var (
	// Command-line flags
	configFile   = flag.String("config", "./configs/config.yaml", "Path to configuration file")
	sourcesFile  = flag.String("sources", "./configs/sources.yaml", "Path to sources configuration file")
	outputFile   = flag.String("output", "", "Output file path")
	outputFormat = flag.String("format", "", "Output format (json, csv, html, xml)")
	logLevel     = flag.String("log-level", "", "Log level (debug, info, warn, error)")
	enableWeb    = flag.Bool("web", false, "Enable web interface")
	webPort      = flag.Int("port", 0, "Web/API server port")
	enableAPI    = flag.Bool("api", false, "Enable API server")
	sourceFilter = flag.String("sources-filter", "", "List of sources names to process")
	version      = flag.Bool("version", false, "Show version information")
)

// Version information (set during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Show version and exit if requested
	if *version {
		fmt.Printf("Web Content Aggregator %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Go version: %s\n", runtime.Version())
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configFile, *sourcesFile)

	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command-line interface if provided
	applyCommandLineOverrides(cfg)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandling(cancel)

	// Create and start the coordinator
	coord, err := coordinator.New(cfg)

	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}

	// Create the aggregator
	agg, err := aggregator.New(cfg, coord)

	if err != nil {
		log.Fatalf("Failed to create aggregator: %v", err)
	}

	// Start web/API servers if enabled
	if cfg.Web.Enabled {
		go startWebServer(cfg, agg)
	}

	if cfg.API.Enabled {
		go startAPIServer(cfg, agg)
	}

	// Run the aggregation process
	startTime := time.Now()
	log.Println("Starting content aggregation...")

	results, err := agg.Run(ctx)

	if err != nil {
		log.Fatalf("Aggregation failed: %v", err)
	}

	// Output the results
	if err := outputResults(cfg, results); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}

	elaspedTime := time.Since(startTime)

	log.Fatalf("Aggregation completed in %v. Processed %d sources, retrieved %d items.",
		elaspedTime, len(cfg.Sources.Sources), len(results))
}

// Applies command-line flag overrides to the configuration
func applyCommandLineOverrides(cfgs *config.Config) {
	// Output file override
	if *outputFile != "" {
		cfgs.App.Output.Destination = "file"
		cfgs.App.Output.FilePath = "outputFile"
	}
}
