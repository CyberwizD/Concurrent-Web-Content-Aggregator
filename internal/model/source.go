package model

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Source represents a web source configuration to fetch content from
type Source struct {
	// Basic information
	Name    string `yaml:"name"`    // Name of the source
	URL     string `yaml:"url"`     // URL of the source
	Enabled bool   `yaml:"enabled"` // Whether the source is enabled

	// Rate limiting settings
	RateLimit struct {
		RequestsPerMinute int  `yaml:"requests_per_minute"` // Maximun requests per minute
		RespectRobotsTxt  bool `yaml:"respect_robots_txt"`  // Whether to respect robots.txt
	} `yaml:"rate_limit"`

	// Parser settings
	Parser string `yaml:"parser"` // Parser type (html, json, xml, rss)

	// Content extraction settings for HTML parsers
	Selector map[string]string `yaml:"selector"` // CSS selectors for HTML parsing

	// Header settings
	Headers map[string]string `yaml:"headers"` // Additional headers to include in requests

	// Content mapping for structured data (JSON, XML)
	Mapping map[string]string `yaml:"mappings"` // Field mappings for structured data

	// Pagnination settings
	Pagination struct {
		Enabled   bool   `yaml:"enabled"`    // Whether pagination is enabled
		StartPage int    `yaml:"start_page"` // First page number
		MaxPages  int    `yaml:"max_pages"`  // Maximum number of pages to fetch
		ParamName string `yaml:"param_name"` // URL parameter for pagination
	} `yaml:"pagination"`

	// Sitemap settings
	Sitemap struct {
		Enabled    bool   `yaml:"enabled"`     // Whether to use sitemap
		ProcessAll bool   `yaml:"process_all"` // Process all URLs or filter
		MaxURLs    int    `yaml:"max_urls"`    // Maximum number of URLs to process
		Pattern    string `yaml:"pattern"`     // Regex pattern for URL filtering
	} `yaml:"sitemap"`
}

// ParseURL parses the source URL into a url.URL struct
func ParseURL(rawURL string) (*url.URL, error) {
	// Support for template placeholders in URLs
	if strings.Contains(rawURL, "${") {
		// For basic validation, replace placeholders with dummy values
		tempURL := rawURL
		tempURL = strings.ReplaceAll(tempURL, "${page}", "1")
		tempURL = strings.ReplaceAll(tempURL, "${date}", time.Now().Format("2006-01-02"))

		return url.Parse(tempURL)
	}

	return url.Parse(rawURL)
}

// GetURLWithPage returns the URL with the page parameter for pagination
func (s *Source) GetURLWithPage(page int) string {
	if !s.Pagination.Enabled {
		return s.URL
	}

	// Replace the page placeholder if present
	if strings.Contains(s.URL, "${page}") {
		return strings.ReplaceAll(s.URL, "${page}", fmt.Sprintf("%d", page))
	}

	// Otherwise, append/replace query parameter
	baseURL, err := url.Parse(s.URL)

	if err != nil {
		// Fallback to simple replacement if URL parsing fails
		return s.URL + "&" + s.Pagination.ParamName + "=" + fmt.Sprintf("%d", page)
	}

	query := baseURL.Query()
	query.Set(s.Pagination.ParamName, fmt.Sprintf("%d", page))
	baseURL.RawQuery = query.Encode()

	return baseURL.String()
}

// Validate checks if the source configuration is valid
func (s *Source) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("source name is required")
	}

	if s.URL == "" {
		return fmt.Errorf("source URL is required")
	}

	// Check if URL is valid, allowing for template placeholders
	_, err := ParseURL(s.URL)

	if err != nil {
		return fmt.Errorf("invalid URL '%s': %w", s.URL, err)
	}

	// Check if parser is specified
	if s.Parser == "" {
		return fmt.Errorf("parser type is required")
	}

	// Validate parser-specific settings
	switch s.Parser {
	case "html":
		if len(s.Selector) == 0 {
			return fmt.Errorf("HTML parser requires at least one selector")
		}
	case "json", "xml", "rss":
		if len(s.Mapping) == 0 {
			return fmt.Errorf("%s parser requires at least one mapping", strings.ToUpper(s.Parser))
		}
	default:
		return fmt.Errorf("unsupported parser type '%s'", s.Parser)
	}

	return nil
}
