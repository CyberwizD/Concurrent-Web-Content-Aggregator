package coordinator

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/fetcher"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/model"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/parser"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

var (
	ErrInvalidConfig      = fmt.Errorf("invalid configuration")
	ErrInvalidCoordinator = fmt.Errorf("invalid coordinator")
)

// Coordinator is the central component that manages concurrency and orchestrates
// the fetching and parsing processes.
type Coordinator struct {
	// Configuration
	config *config.Config

	// Worker pools
	fetcherPool *WorkerPool
	parserPool  *WorkerPool

	// Fetcher and parsers
	fetcher *fetcher.Fetcher
	parsers map[string]parser.Parser

	// Channels for data flow
	fetchJobs    chan *model.FetchJob
	fetchResults chan *model.FetchResult
	parseJobs    chan *model.ParseJob
	parseResults chan *model.ParseResult

	// Statistics
	stats *Stats

	// Mutex for stats updates
	mu sync.Mutexes
}

// Stats tracks statistics about the aggregation process
type Stats struct {
	TotalSources      int
	ProcessedSources  int
	SuccessfulFetches int
	FailedFetches     int
	SuccessfulParses  int
	FailedParses      int
	StartTime         time.Time
	EndTime           time.Time
}

// New creates a new coordinator with the provided configuration
func New(cfg *config.Config) (*Coordinator, error) {
	if cfg == nil {
		return nil, errors.New("configuration is required")
	}

	// Create fetcher
	f, err := fetcher.New(cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to create fetcher: %v", err)
	}

	// Create parsers for each type
	parsers := make(map[string]parser.Parser)

	for _, parseType := range []string{"html", "json", "xml", "rss"} {
		p, err := parser.Get(parseType)

		if err != nil {
			return nil, fmt.Errorf("failed to create parser for type %s: %v", parserType, err)
		}

		parsers[parseType] = p
	}

	// Create data channels
	fetchJobChan := make(chan *model.FetchJob, cfg.App.Concurrency.MaxFetchers)
	fetchResultChan := make(chan *model.FetchResult, cfg.App.Concurrency.MaxFetchers)
	parseJobChan := make(chan *model.ParseJob, cfg.App.Concurrency.MaxParsers)
	parseResultChan := make(chan *model.ParseResult, cfg.App.Concurrency.MaxParsers)

	// Create worker pools
	fetcherPool := NewWorkerPool(cfg.App.Concurrency.MaxFetchers, "fetcher")
	parserPool := NewWorkerPool(cfg.App.Concurrency.MaxParsers, "parser")

	return &Coordinator{
		config:       cfg,
		fetcher:      f,
		parsers:      parsers,
		fetcherPool:  fetcherPool,
		parserPool:   parserPool,
		fetchJobs:    fetchJobChan,
		fetchResults: fetchResultChan,
		parseJobs:    parseJobChan,
		parseResults: parseResultChan,
		stats: &Stats{
			TotalSources: len(cfg.Sources.Sources),
			StartTime:    time.Now(),
		},
	}, nil
}
