package fetcher

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/model"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
	"github.com/temoto/robotstxt"
)

// Fetcher is responsible for fetching content from web sources
type Fetcher struct {
	client      *http.Client
	config      *config.Config
	rateLimiter *RateLimiter
	robotsCache map[string]*robotstxt.RobotsData
	robotsMu    sync.RWMutex
}

// Create content object
type content struct {
	Source      *model.Source
	URL         *url.URL
	Body        []byte
	ContentType string
	StatusCode  int
	Headers     http.Header
	FetchedAt   time.Time
}

// New creates a new Fetcher with the provided configuration
func New(cfg *config.Config) (*Fetcher, error) {
	// Create HTTP client with configured timeouts
	client := &http.Client{
		Timeout: time.Duration(cfg.App.Timeouts.Request),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.App.HTTP.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", cfg.App.HTTP.MaxRedirects)
			}
			return nil
		},
	}

	// Create rate limiter
	limiter := NewRateLimiter(cfg)

	return &Fetcher{
		client:      client,
		config:      cfg,
		rateLimiter: limiter,
		robotsCache: make(map[string]*robotstxt.RobotsData),
	}, nil
}

// Fetch retrieves content from the specified source
func (f *Fetcher) Fetch(ctx context.Context, source *model.Source) (*model.Content, error) {
	// Parse URL
	parsedURL, err := url.Parse(source.URl)

	if err != nil {
		return nil, fmt.Errorf("invalid URL '%s': %w", source.URL, err)
	}

	// Check robots.txt if configured
	if source.RateLimit.RespectRobotsTxt {
		allowed, err := f.checkRobotsTxt(ctx, parsedURL, source)

		if err != nil {
			log.Printf("Warning: Error checking robots.txt for %s: %v", parsedURL.Host, err)
			// Continue anyway since it's just a warning
		} else if !allowed {
			return nil, fmt.Errorf("URL '%s' disallowed by robots.txt", sources.URL)
		}
	}

	// Apply rate limiting
	err = f.rateLimiter.Wait(ctx, parsedURL.Host)

	if err != nil {
		return nil, fmt.Errorf("rate limiting error: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", f.config.App.UserAgent)

	for key, value := range source.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	log.Printf("Fetching %s", source.URL)

	resp, err := f.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errof("failed to read response body: %w", err)
	}

	return &content{
		Source:      source,
		URL:         source.URL,
		Body:        body,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		Headers:     resp.Header,
		FetchedAt:   time.Now(),
	}, nil
}

// checkRobotsTxt verifies if the URL is allowed by the site's robots.txt
func (f *Fetcher) checkRobotsTxt(ctx context.Context, parsedURL *url.URL, source *model.Source) (bool, error) {
	host := parsedURL.Host
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, host)

	// Check cache first
	f.robotsMu.RLock()
	robotsData, exists := f.robotsCache[host]
}
