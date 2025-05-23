package fetcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/model"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

// RateLimiter controls request frequency to prevent overwhelming websites
type RateLimiter struct {
	// Map of domain -> token bucket
	buckets map[string]*TokenBucket

	// Default requests per minute for domains without specific config
	defaultRPM int

	// Mutex for safe concurrent access to buckets map
	mu sync.Mutex

	// Configuration
	config *config.Config
}

// TokenBucket implements a token bucket rate limiting algorithm
type TokenBucket struct {
	// Maximun number of tokens the bucket can hold
	capacity int

	// Current number of tokens in the bucket
	tokens int

	// Rate at which tokens are added to the bucket (tokens per second)
	rate float64

	// Last time tokens were added to the bucket
	lastRefill time.Time

	// Mutex for concurrent access
	mu sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.Config) *RateLimiter {
	// Default to 30 requests per minute if not specified
	defaultRPM := 30

	return &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		defaultRPM: defaultRPM,
		config:     cfg,
	}
}

// Wait blocks until a request can be made for the given domain
func (rl *RateLimiter) Wait(ctx context.Context, domain string) error {
	bucket := rl.getBucket(domain)

	for {
		// Check if we can take a token
		if bucket.Take() {
			return nil
		}

		// No tokens available, wait a bit and try again
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for rate limit: %w", ctx.Err())
		case <-time.After(50 * time.Millisecond):
			// Try again
		}
	}
}

// getBucket returns the token bucket for a domain, creating it if needed
func (rl *RateLimiter) getBucket(domain string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[domain]

	if !exists {
		// Try to find domain-specific rate limit to config
		rpm := rl.getDomainRateLimit(domain)

		// Create a new bucket
		bucket = &TokenBucket{
			capacity:   rpm,
			tokens:     rpm,                 // Start full
			rate:       float64(rpm) / 60.0, // Convert RPM to tokens per second
			lastRefill: time.Now(),
		}

		rl.buckets[domain] = bucket
		log.Printf("Create rate limiter for %s: %d requests per minute", domain, rpm)
	}

	return bucket
}

// getDomainRateLimit returns the rate limit for a specific domain
func (rl *RateLimiter) getDomainRateLimit(domain string) int {
	// check if we have a specific limit for this domain
	for _, source := range rl.config.Sources.Sources {
		sourceURL, err := model.ParseURL(source.URL)

		if err != nil {
			continue
		}

		if sourceURL.Host == domain && source.RateLimit.RequestsPerMinute > 0 {
			return source.RateLimit.RequestsPerMinute
		}
	}

	// Return default rate limit
	return rl.defaultRPM
}

// Take attempts to take a token from the bucket
// Returns true if a token was taken, false otherwise
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elasped since last refill
	tb.refill()

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false

}

// Refill adds tokens to the bucket based on the rate and time elapsed
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// Calculate the number of tokens to add
	tokensToAdd := int(elapsed * tb.rate)

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd

		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}

		tb.lastRefill = now
	}
}
