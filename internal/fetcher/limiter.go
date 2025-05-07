package fetcher

import (
	"sync"

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

type TokenBucket struct {
}

func NewRateLimiter(cfg *config.Config) (*RateLimiter, error) {

}
