package fetcher

import (
	"sync"
	"time"

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
