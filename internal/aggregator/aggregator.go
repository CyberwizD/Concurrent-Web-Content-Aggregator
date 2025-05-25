package aggregator

import (
	"context"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/coordinator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

type Aggregator struct {
	Coordinator *coordinator.Coordinator
	Config      *config.Config
}

type Result struct {
	// Add fields to store results of the aggregation
	AggregatedContent []string
	Statistics        map[string]int
	Errors            []error
}

// New creates a new Aggregator instance with the provided configuration and coordinator.
func New(cfg *config.Config, coord *coordinator.Coordinator) (*Aggregator, error) {
	if cfg == nil {
		return nil, coordinator.ErrInvalidConfig
	}

	if coord == nil {
		return nil, coordinator.ErrInvalidCoordinator
	}

	return &Aggregator{
		Config:      cfg,
		Coordinator: coord,
	}, nil
}

func (a *Aggregator) Run(ctx context.Context) ([]Result, error) {
	// Use the existing coordinator instance
	if err := a.Coordinator.Start(ctx); err != nil {
		return nil, err
	}

	// Start the coordinator
	if err := a.Coordinator.Start(ctx); err != nil {
		return nil, err
	}

	// Wait for the coordinator to finish
	if err := a.Coordinator.Wait(); err != nil {
		return nil, err
	}

	// Return the results
	return []Result{}, nil
}

// WriteResultsToFile writes the aggregation results to a file in the specified format
func WriteResultsToFile(filePath string, results []Result, format string) error {
	// TODO: Implement file writing logic based on the format
	return nil
}

func WriteResultsToStdout(results []Result, format string) error {
	// TODO: Implement stdout writing logic based on the format
	return nil
}
