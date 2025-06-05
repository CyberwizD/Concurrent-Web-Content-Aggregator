package aggregator

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"os"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/coordinator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

// ErrInvalidAggregator is returned when an invalid aggregator is provided
var ErrInvalidAggregator = errors.New("aggregator cannot be nil")

type Aggregator struct {
	Coordinator *coordinator.Coordinator
	Config      *config.Config
	Content     string
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
	f, err := os.Create(filePath)

	if err != nil {
		return err
	}

	defer f.Close()

	switch format {
	case "json":
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	case "csv":
		writer := csv.NewWriter(f)
		defer writer.Flush()
		writer.Write([]string{"Content", "Error Count"})

		for _, res := range results {
			content := ""
			if len(res.AggregatedContent) > 0 {
				content = res.AggregatedContent[0] // simplified
			}
			writer.Write([]string{content, fmt.Sprintf("%d", len(res.Errors))})
		}
	case "xml":
		encoder := xml.NewEncoder(f)
		return encoder.Encode(results)
	case "html":
		tmpl := `<html><body><h1>Aggregated Results</h1>{{range .}}<p>{{index .AggregatedContent 0}}</p>{{end}}</body></html>`
		t := template.Must(template.New("results").Parse(tmpl))
		return t.Execute(f, results)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

func WriteResultsToStdout(results []Result, format string) error {
	// TODO: Implement stdout writing logic based on the format
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()
		writer.Write([]string{"Content", "Error Count"})
		for _, res := range results {
			content := ""
			if len(res.AggregatedContent) > 0 {
				content = res.AggregatedContent[0]
			}
			writer.Write([]string{content, fmt.Sprintf("%d", len(res.Errors))})
		}
	case "xml":
		encoder := xml.NewEncoder(os.Stdout)
		return encoder.Encode(results)
	case "html":
		tmpl := `<html><body><h1>Aggregated Results</h1>{{range .}}<p>{{index .AggregatedContent 0}}</p>{{end}}</body></html>`
		t := template.Must(template.New("results").Parse(tmpl))
		return t.Execute(os.Stdout, results)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}
