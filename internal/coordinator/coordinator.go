package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	mu sync.Mutex
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
			return nil, fmt.Errorf("failed to create parser for type %s: %v", parseType, err)
		}

		parsers[parseType] = *p
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

// Start initializes and starts all worker pools and processing pipelines
func (c *Coordinator) Start(ctx context.Context) error {
	// Start fetcher workers
	c.fetcherPool.Start(ctx, func(id int, ctx context.Context) {
		c.fetchWorker(ctx, id)
	})

	// Start Parser workers
	c.parserPool.Start(ctx, func(id int, ctx context.Context) {
		c.parseWorker(ctx, id)
	})

	log.Printf("Coordinator started with %d fetchers and %d parsers",
		c.config.App.Concurrency.MaxFetchers,
		c.config.App.Concurrency.MaxParsers,
	)

	return nil
}

// Stop gracefully shuts down all worker pools
func (c *Coordinator) Stop() {
	// Stop worker pools
	c.fetcherPool.Stop()
	c.parserPool.Stop()

	// Close channel
	close(c.fetchJobs)
	close(c.parseJobs)

	// Update end time
	c.mu.Lock()
	c.stats.EndTime = time.Now()
	c.mu.Unlock()

	log.Printf("Coordinator stopped")
}

// Wait blocks until all fetch and parse jobs are completed
func (c *Coordinator) Wait() error {
	// Wait for all fetch jobs to complete
	go func() {
		for job := range c.fetchResults {
			c.mu.Lock()
			if job.Error != nil {
				c.stats.FailedFetches++
				log.Printf("Fetch failed for %s: %v", job.Source.URL, job.Error)
			} else {
				c.stats.SuccessfulFetches++
				c.stats.ProcessedSources++
			}
			c.mu.Unlock()
		}
	}()

	// Wait for all parse jobs to complete
	go func() {
		for job := range c.parseResults {
			c.mu.Lock()
			if job.Error != nil {
				c.stats.FailedParses++
				log.Printf("Parse failed for %s: %v", job.Source.Name, job.Error)
			} else {
				c.stats.SuccessfulParses++
			}
			c.mu.Unlock()
		}
	}()

	return nil
}

// GetStats returns a copy of the current statistics
func (c *Coordinator) GetStats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to avoid race condition
	statsCopy := *c.stats
	return statsCopy
}

// SubmitFetchJob submits a source to be fetched
func (c *Coordinator) SubmitFetchJob(source *model.Source) {
	job := &model.FetchJob{
		Source:      source,
		SubmittedAt: time.Now(),
	}

	c.fetchJobs <- job
}

// GetFetchResults returns a channel for receiving fetch results
func (c *Coordinator) GetFetchResults() <-chan *model.FetchResult {
	return c.fetchResults
}

// GetParseResults returns a channel for receiving parse results
func (c *Coordinator) GetParseResults() <-chan *model.ParseResult {
	return c.parseResults
}

// fetchWorker processes fetch jobs from the fetch jobs channel
func (c *Coordinator) fetchWorker(ctx context.Context, workerID int) {
	log.Printf("Fetch worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Fetch worker %d stopping: context cancelled", workerID)
			return
		case job, ok := <-c.fetchJobs:
			if !ok {
				log.Printf("Fetch worker %d stopping channel closed", workerID)
				return
			}

			// Process the fetch job
			result := c.processFetchJob(ctx, job, workerID)

			// Send the result
			select {
			case c.fetchResults <- result:
				// Result sent successfully
			case <-ctx.Done():
				log.Printf("Fetch worker %d: context cancelled while sending result", workerID)
				return
			}

			// If fetch was successful, submit for parsing
			if result.Error == nil && result.Content != nil {
				parseJob := &model.ParseJob{
					Source:      job.Source,
					Content:     result.Content,
					SubmittedAt: time.Now(),
				}

				select {
				case c.parseJobs <- parseJob:
					// Parse job submitted successfully
				case <-ctx.Done():
					log.Printf("Fetch worker %d: context cancelled while submitting parse job", workerID)
					return
				}
			}
		}
	}
}

// processFetchJob handles the actual fetching of content
func (c *Coordinator) processFetchJob(ctx context.Context, job *model.FetchJob, workerID int) *model.FetchResult {
	log.Printf("worker %d fetching from %s", workerID, job.Source.URL)

	// Create fetch-specific context with timeout
	fetchCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.App.Timeouts.Request))
	defer cancel()

	// Fetch the content
	content, err := c.fetcher.Fetch(fetchCtx, job.Source)

	// Create result
	result := &model.FetchResult{
		Source:      job.Source,
		Content:     content,
		FetchedAt:   time.Now(),
		ProcessedBy: workerID,
		Error:       err,
	}

	// Update stats
	c.mu.Lock()
	if err != nil {
		c.stats.FailedFetches++
	} else {
		c.stats.SuccessfulFetches++
	}

	c.stats.ProcessedSources = c.stats.SuccessfulFetches + c.stats.FailedFetches
	c.mu.Unlock()

	return result
}

// parseWorker processes parse jobs from the parse jobs channel
func (c *Coordinator) parseWorker(ctx context.Context, workerID int) {
	log.Printf("Parse worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Parse worker %d stopping: context cancelled", workerID)
			return
		case job, ok := <-c.parseJobs:
			if !ok {
				log.Printf("Parse worker %d stopping: channel closed", workerID)
				return
			}

			// Process the parse job
			result := c.processParseJob(ctx, job, workerID)

			// Send the result
			select {
			case c.parseResults <- result:
				// Result sent successfully
			case <-ctx.Done():
				log.Printf("Parse worker %d: context cancelled while sending result", workerID)
				return
			}
		}
	}
}

// processParseJob handles the actual parsing of content
func (c *Coordinator) processParseJob(ctx context.Context, job *model.ParseJob, workerID int) *model.ParseResult {
	source := job.Source
	log.Printf("worker %d parsing content from %s using %s parser", workerID, source.Name, source.Parser)

	// Get the appropriate parser
	parser, exists := c.parsers[source.Parser]
	if !exists {
		err := fmt.Errorf("no parser available for type: %s", source.Parser)

		// Update stats
		c.mu.Lock()
		c.stats.FailedParses++
		c.mu.Unlock()

		return &model.ParseResult{
			Source:      source,
			ParsedAt:    time.Now(),
			ProcessedBy: workerID,
			Error:       err,
		}
	}

	// Parse the content
	items, err := parser.Parse(ctx, job.Content, source)

	// Create result
	result := &model.ParseResult{
		Source:      source,
		Items:       items,
		ParsedAt:    time.Now(),
		ProcessedBy: workerID,
		Error:       err,
	}

	// Update stats
	c.mu.Lock()
	if err != nil {
		c.stats.FailedParses++
	} else {
		c.stats.SuccessfulParses++
	}
	c.mu.Unlock()

	return result
}
