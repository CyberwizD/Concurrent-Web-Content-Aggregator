package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
func applyCommandLineOverrides(cfg *config.Config) {
	// Output file override
	if *outputFile != "" {
		cfg.App.Output.Destination = "file"
		cfg.App.Output.FilePath = *outputFile
	}

	// Output format override
	if *outputFormat != "" {
		cfg.App.Output.Format = *outputFormat
	}

	// Log level override
	if *logLevel != "" {
		switch *logLevel {
		case "debug":
			cfg.App.Log.Level = config.LogLevelDebug
		case "info":
			cfg.App.Log.Level = config.LogLevelInfo
		case "warn":
			cfg.App.Log.Level = config.LogLevelWarn
		case "error":
			cfg.App.Log.Level = config.LogLevelError
		default:
			log.Fatalf("Invalid log level: %s", *logLevel)
		}
	}

	// Web server override
	if *enableWeb {
		cfg.Web.Enabled = true

		if *webPort > 0 {
			cfg.Web.Port = *webPort
		}
	}

	// API server override
	if *enableAPI {
		cfg.API.Enabled = true

		if *webPort > 0 {
			cfg.API.Port = *webPort
		}
	}

	// Source filter override
	if *sourceFilter != "" {
		// Implementation depends on how you want to filter sources
		// For now, just log that we'd filter
		log.Printf("Filtering sources: %s", *sourceFilter)
		// TODO: Implement source filtering
	}
}

// setupSignalHandling configures handling of OS signals for graceful shutdown
func setupSignalHandling(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-c
		log.Printf("Received signal: %s. Shutting down...", sig)

		// Call the cancel function to stop the context
		cancel()

		// Allow some time for graceful shutdown before forcing exit
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}()
}

// startWebServer initializes and starts the web interface server
func startWebServer(cfg *config.Config, agg *aggregator.Aggregator) {
	log.Printf("Starting web server on port %d...", cfg.Web.Port)
	// TODO: Implement web server
	log.Printf("Web server started on http://%s:%d", cfg.Web.Host, cfg.Web.Port)
}

// startAPIServer initializes and startss the API server
func startAPIServer(cfg *config.Config, agg *aggregator.Aggregator) {
	log.Printf("Starting API server on port %d...", cfg.API.Port)
	// TODO: Implement API server
	log.Printf("API server started on http://%s:%d", cfg.API.Host, cfg.API.Port)
}

// outputResults handles writing the aggregation results to the configured destination
func outputResults(cfg *config.Config, results []aggregator.Result) error {
	switch cfg.App.Output.Destination {
	case "file":
		if cfg.App.Output.FilePath == "" {
			return fmt.Errorf("output file path is not set")
		}
		return aggregator.WriteResultsToFile(cfg.App.Output.FilePath, results, cfg.App.Output.Format)
	case "stdout":
		return aggregator.WriteResultsToStdout(results, cfg.App.Output.Format)
	default:
		return fmt.Errorf("unknown output destination: %s", cfg.App.Output.Destination)
	}
}
