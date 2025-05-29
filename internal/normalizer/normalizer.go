package normalizer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/aggregator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/parser"
)

// Normalizer transforms raw content into a normalized format.
// before sending it to the result aggregator.
type Normalizer struct {
	aggregator   *aggregator.Aggregator
	normalizeMap map[string]NormalizerFunc
	mu           sync.Mutex
}

// NormalizerFunc is a function that normalizes data for a specific content type
type NormalizerFunc func(*parser.Parser) (*aggregator.Aggregator, error)

// New creates a new Normalizer instance with the provided aggregator.
func New(agg *aggregator.Aggregator) (*Normalizer, error) {
	if agg == nil {
		return nil, fmt.Errorf("invalid aggregator: %w", aggregator.ErrInvalidAggregator)
	}

	n := &Normalizer{
		aggregator:   agg,
		normalizeMap: make(map[string]NormalizerFunc),
	}

	// Register default normalizers
	n.RegisterNormalizer("text", normalizeText)
	n.RegisterNormalizer("html", normalizeHTML)
	n.RegisterNormalizer("json", normalizeJSON) // Example for JSON, can be replaced with actual JSON normalization

	return n, nil
}

// Normalize processes the content using the appropriate normalizer based on the content type.
func (n *Normalizer) Normalize(p *parser.Parser) (*aggregator.Aggregator, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	contentType := p.GetContentType()
	if normalizeFunc, exists := n.normalizeMap[contentType]; exists {
		log.Printf("Normalizing content of type '%s'", contentType)
		startTime := time.Now()
		normalizedContent, err := normalizeFunc(p)
		if err != nil {
			return nil, fmt.Errorf("normalization failed for content type '%s': %w", contentType, err)
		}
		duration := time.Since(startTime)
		log.Printf("Normalization completed for content type '%s' in %s", contentType, duration)
		return normalizedContent, nil
	}

	log.Printf("No normalizer registered for content type '%s', returning raw content", contentType)
	return &aggregator.Aggregator{
		Content: p.GetContent(),
	}, nil
}

func (n *Normalizer) RegisterNormalizer(contentType string, normalizeFunc NormalizerFunc) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.normalizeMap[contentType]; exists {
		log.Printf("Normalizer for content type '%s' already exists, overwriting", contentType)
	}
	n.normalizeMap[contentType] = normalizeFunc
}

func normalizeHTML(p *parser.Parser) (*aggregator.Aggregator, error) {
	// Example normalization logic for HTML content
	content := p.GetContent()
	if content == "" {
		return nil, fmt.Errorf("no content to normalize")
	}

	// Simulate normalization process
	normalizedContent := strings.TrimSpace(content)
	log.Printf("Normalized HTML content: %s", normalizedContent)

	// Create a new aggregator instance with the normalized content
	return &aggregator.Aggregator{
		Content: normalizedContent,
	}, nil
}

func normalizeText(p *parser.Parser) (*aggregator.Aggregator, error) {
	// Example normalization logic for text content
	content := p.GetContent()
	if content == "" {
		return nil, fmt.Errorf("no content to normalize")
	}

	// Simulate normalization process
	normalizedContent := strings.TrimSpace(content)
	log.Printf("Normalized text content: %s", normalizedContent)

	// Create a new aggregator instance with the normalized content
	return &aggregator.Aggregator{
		Content: normalizedContent,
	}, nil
}

func normalizeJSON(p *parser.Parser) (*aggregator.Aggregator, error) {
	// Example normalization logic for JSON content
	content := p.GetContent()
	if content == "" {
		return nil, fmt.Errorf("no content to normalize")
	}

	// Simulate normalization process
	normalizedContent := strings.TrimSpace(content)
	log.Printf("Normalized JSON content: %s", normalizedContent)

	// Create a new aggregator instance with the normalized content
	return &aggregator.Aggregator{
		Content: normalizedContent,
	}, nil
}
