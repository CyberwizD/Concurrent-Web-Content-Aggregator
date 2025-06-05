package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RobotsData represents parsed robots.txt rules
type RobotsData struct {
	Rules      map[string][]Rule
	Sitemaps   []string
	Crawldelay time.Duration
	Expires    time.Time
}

// Rule represents a single robots.txt rule
type Rule struct {
	Path  string
	Allow bool
}

// RobotsCache manages cached robots.txt data for various domains
type RobotsCache struct {
	cache          map[string]*RobotsData
	client         *http.Client
	userAgent      string
	expiration     time.Duration
	respectNoIndex bool
	mu             sync.RWMutex
}

// RobotsCacheOptions configures the robots.txt cache
type RobotsCacheOptions struct {
	Client         *http.Client
	UserAgent      string
	Expiration     time.Duration
	RespectNoIndex bool
}

// NewRobotsCache create a new robots.txt cache
func NewRobotsCache(opts RobotsCacheOptions) *RobotsCache {
	// set options
	client := opts.Client
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = "WebAggregator/1.0"
	}

	expiration := opts.Expiration
	if expiration < 0 {
		expiration = 24 * time.Hour
	}

	return &RobotsCache{
		cache:          make(map[string]*RobotsData),
		client:         client,
		userAgent:      userAgent,
		expiration:     expiration,
		respectNoIndex: opts.RespectNoIndex,
	}
}

// IsAllowed checks if the given URL is allowed to be crawled
func (r *RobotsCache) IsAllowed(parsedURL *url.URL, agent string) (bool, error) {
	if parsedURL == nil {
		return false, fmt.Errorf("no URL provided")
	}

	// Extract the host
	host := parsedURL.Host

	// Check if there's need to refresh the cache
	robotsData, err := r.getRobotsData(parsedURL)

	if err != nil {
		// Crawl and log the error
		log.Printf("Warning: Couldn't fetch robots.txt for %s: %v", host, err)
		return true, nil
	}

	// Find the rules for this user agent
	var rules []Rule

	agentRules, exists := robotsData.Rules[agent]

	if exists {
		rules = agentRules
	} else {
		// Find the rules for the wildcard agent
		agentRules, exists = robotsData.Rules["*"]
		if exists {
			rules = agentRules
		}
	}

	// if no rules were found, allow by default
	if len(rules) == 0 {
		return true, nil
	}

	// Get the path to check
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	// Check the rules
	return r.checkRules(rules, path), nil

}

// getRobotsData fetches and parses robots.txt data for a domain
func (r *RobotsCache) getRobotsData(parsedURL *url.URL) (*RobotsData, error) {
	host := parsedURL.Host

	// Check if we have a valid cached entry
	r.mu.RLock()
	cachedData, exists := r.cache[host]
	r.mu.RUnlock()

	if exists && time.Now().Before(cachedData.Expires) {
		return cachedData, nil
	}

	// Need to fetch/refresh the robots.txt
	robotsURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   host,
		Path:   "/robots.txt",
	}

	// Fetch robots.txt
	req, err := http.NewRequest("GET", robotsURL.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", r.userAgent)

	resp, err := r.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch robots.txt: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If not found or error, use an empty robots.txt
		if resp.StatusCode == http.StatusNotFound {
			// Create an empty robots data with rules allowing everything
			emptyData := &RobotsData{
				Rules:   make(map[string][]Rule),
				Expires: time.Now().Add(r.expiration),
			}

			r.mu.Lock()
			r.cache[host] = emptyData
			r.mu.Unlock()

			return emptyData, nil
		}

		return nil, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	// Read and parse teh robots.txt content
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	robotsData, err := r.parseRobotsTxt(string(body))

	if err != nil {
		return nil, fmt.Errorf("failed to parse robots.txt: %w", err)
	}

	// Set expiration time
	robotsData.Expires = time.Now().Add(r.expiration)

	// Cache the result
	r.mu.Lock()
	r.cache[host] = robotsData
	r.mu.Unlock()

	return robotsData, nil
}

// parseRobotsTxt parses robots.txt content and returns structured data
func (r *RobotsCache) parseRobotsTxt(content string) (*RobotsData, error) {
	data := &RobotsData{
		Rules:    make(map[string][]Rule),
		Sitemaps: []string{},
	}

	var currentAgent string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Remove comments and trim whitespace
		if idx := strings.IndexByte(line, '#'); idx != -1 {
			line = line[:idx]
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Split the line into parts
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch field {
		case "user-agent":
			currentAgent = value
			if _, exits := data.Rules[currentAgent]; !exits {
				data.Rules[currentAgent] = []Rule{}
			}
		case "disallow":
			if currentAgent != "" && value != "" {
				data.Rules[currentAgent] = append(data.Rules[currentAgent], Rule{
					Path:  value,
					Allow: true,
				})
			}
		case "sitemap":
			if value != "" {
				data.Sitemaps = append(data.Sitemaps, value)
			}
		case "crawl-delay":
			// parse crawl delay if present
			var delay int
			if _, err := fmt.Sscanf(value, "%d", &delay); err == nil && delay > 0 {
				data.Crawldelay = time.Duration(delay) * time.Second
			}
		}
	}

	return data, nil
}

// checkRules evaluates if a path is allowed based on the rules
func (r *RobotsCache) checkRules(rules []Rule, path string) bool {
	// Default to allowing if no rules
	if len(rules) == 0 {
		return true
	}

	// Find the most specific matching rule
	var longestMatch *Rule
	longestMatchLength := -1

	for i := range rules {
		rule := &rules[i]
		rulePath := rule.Path

		// Handle wildcards
		if strings.HasSuffix(rulePath, "*") {
			prefix := rulePath[:len(rulePath)-1]

			if strings.HasPrefix(path, prefix) && len(prefix) > longestMatchLength {
				longestMatch = rule
				longestMatchLength = len(prefix)
			}
		} else {
			// Exact match or path prefix
			if (path == rulePath || strings.HasPrefix(path, rulePath+"/")) &&
				len(rulePath) > longestMatchLength {
				longestMatch = rule
				longestMatchLength = len(rulePath)
			}
		}
	}

	// If no rule matched, allow by default
	if longestMatch == nil {
		return true
	}

	// Return whether the matching rule allows or disallows
	return longestMatch.Allow
}

// GetSitemaps returns the list of sitemaps for a domain
func (r *RobotsCache) GetSitemaps(parsedURL *url.URL) ([]string, error) {
	robotsData, err := r.getRobotsData(parsedURL)
	if err != nil {
		return nil, err
	}

	return robotsData.Sitemaps, nil
}

// GetCrawlDelay returns the crawl delay for a domain and user agent
func (r *RobotsCache) GetCrawlDelay(parsedURL *url.URL, agent string) (time.Duration, error) {
	robotsData, err := r.getRobotsData(parsedURL)
	if err != nil {
		return 0, err
	}

	return robotsData.Crawldelay, nil
}

// ClearCache removes all cached robots.txt data
func (r *RobotsCache) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache = make(map[string]*RobotsData)
}

// RemoveFromCache removes a specific domain from the cache
func (r *RobotsCache) RemoveFromCache(host string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.cache, host)
}
