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
	parsedURL, err := url.Parse(source.URL)

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
			return nil, fmt.Errorf("URL '%s' disallowed by robots.txt", source.URL)
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
	req.Header.Set("User-Agent", f.config.App.HTTP.UserAgent)

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
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &model.Content{
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
	f.robotsMu.RUnlock()

	if !exists {
		// Fetch robots.txt
		req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
		if err != nil {
			return true, fmt.Errorf("failed to create robots.txt request: %w", err)
		}

		// Use source-specific User-Agent if available, otherwise use default
		userAgent := f.config.App.HTTP.UserAgent
		if agent, ok := source.Headers["User-Agent"]; ok && agent != "" {
			userAgent = agent
		}
		req.Header.Set("User-Agent", userAgent)

		// Apply any other source-specific headers that might be relevant
		for key, value := range source.Headers {
			if key != "User-Agent" { // Skip User-Agent as it's already set
				req.Header.Set(key, value)
			}
		}

		// Execute request with rate limiting
		err = f.rateLimiter.Wait(ctx, host)
		if err != nil {
			return true, fmt.Errorf("rate limit check failed for robots.txt: %w", err)
		}

		resp, err := f.client.Do(req)
		if err != nil {
			return true, fmt.Errorf("robots.txt request failed: %w", err)
		}
		defer resp.Body.Close()

		// Check status code - if robots.txt not found or error, allow access
		if resp.StatusCode >= 400 {
			return true, nil
		}

		// Parse robots.txt
		robotsTxt, err := robotstxt.FromResponse(resp)
		if err != nil {
			return true, fmt.Errorf("failed to parse robots.txt: %w", err)
		}

		// Cache the robots.txt data
		f.robotsMu.Lock()
		f.robotsCache[host] = robotsTxt
		robotsData = robotsTxt
		f.robotsMu.Unlock()
	}

	// Check if URL is allowed using source-specific User-Agent
	userAgent := f.config.App.HTTP.UserAgent
	if agent, ok := source.Headers["User-Agent"]; ok && agent != "" {
		userAgent = agent
	}
	group := robotsData.FindGroup(userAgent)
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	allowed := group.Test(path)
	if !allowed {
		return false, fmt.Errorf("URL path '%s' is disallowed by robots.txt for User-Agent '%s'", path, userAgent)
	}

	return true, nil
}
