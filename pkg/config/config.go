// package config

// import (
// 	"fmt"
// )

// type Config struct {
// 	// App configuration
// 	App AppConfig `yaml:"app"`

// 	// Sources configuration
// 	Sources SourcesConfig `yaml:"sources"`

// 	// Fetcher configuration
// 	Fetcher FetcherConfig `yaml:"fetcher"`

// 	// Parser configuration
// 	Parser ParserConfig `yaml:"parser"`

// 	// Aggregator configuration
// 	Aggregator AggregatorConfig `yaml:"aggregator"`

// 	// Output configuration
// 	Output OutputConfig `yaml:"output"`

// 	// Web configuration
// 	Web WebConfig `yaml:"web"`
// 	// Coordinator configuration
// 	// API configuration
// 	API APIConfig `yaml:"api"`
// }

package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main application configuration
type Config struct {
	App     AppConfig     `yaml:"app"`
	Fetcher FetcherConfig `yaml:"fetcher"`
	Parser  ParserConfig  `yaml:"parser"`
	Output  OutputConfig  `yaml:"output"`
	Logging LoggingConfig `yaml:"logging"`
	Sources SourcesConfig `yaml:"sources"`
	Web     WebConfig     `yaml:"web"`
	API     APIConfig     `yaml:"api"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Name        string        `yaml:"name"`
	Environment string        `yaml:"environment"`
	Debug       bool          `yaml:"debug"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxRetries  int           `yaml:"max_retries"`
	Output      OutputConfig  `yaml:"output"`
}

// FetcherConfig contains settings for the fetcher components
type FetcherConfig struct {
	MaxConcurrentWorkers int           `yaml:"max_concurrent_workers"`
	RequestTimeout       time.Duration `yaml:"request_timeout"`
	MaxIdleConns         int           `yaml:"max_idle_conns"`
	MaxConnsPerHost      int           `yaml:"max_conns_per_host"`
	UserAgent            string        `yaml:"user_agent"`
	RespectRobotsTxt     bool          `yaml:"respect_robots_txt"`
	RateLimiting         RateLimit     `yaml:"rate_limiting"`
	RetryPolicy          RetryPolicy   `yaml:"retry_policy"`
}

// RateLimit contains rate limiting configuration
type RateLimit struct {
	Enabled           bool          `yaml:"enabled"`
	RequestsPerMinute int           `yaml:"requests_per_minute"`
	BurstSize         int           `yaml:"burst_size"`
	PerDomain         bool          `yaml:"per_domain"`
	DefaultDelay      time.Duration `yaml:"default_delay"`
}

// RetryPolicy contains retry configuration
type RetryPolicy struct {
	MaxRetries      int           `yaml:"max_retries"`
	InitialDelay    time.Duration `yaml:"initial_delay"`
	MaxDelay        time.Duration `yaml:"max_delay"`
	BackoffFactor   float64       `yaml:"backoff_factor"`
	RetryableErrors []string      `yaml:"retryable_errors"`
}

// ParserConfig contains settings for the parser components
type ParserConfig struct {
	MaxConcurrentWorkers int                    `yaml:"max_concurrent_workers"`
	BufferSize           int                    `yaml:"buffer_size"`
	Timeout              time.Duration          `yaml:"timeout"`
	Selectors            map[string]interface{} `yaml:"selectors"` // CSS selectors for HTML parsing
}

// OutputConfig contains settings for output handling
type OutputConfig struct {
	Format      string          `yaml:"format"`      // json, yaml, csv, xml
	Destination string          `yaml:"destination"` // file, stdout, api
	FilePath    string          `yaml:"file_path"`
	Pretty      bool            `yaml:"pretty"`
	API         APIOutputConfig `yaml:"api"`
	Database    DatabaseConfig  `yaml:"database"`
}

// APIOutputConfig contains API output settings
type APIOutputConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	BasePath string `yaml:"base_path"`
}

// DatabaseConfig contains database settings for optional storage
type DatabaseConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Driver   string `yaml:"driver"` // sqlite, postgres, mysql
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`  // debug, info, warn, error
	Format     string `yaml:"format"` // text, json
	Output     string `yaml:"output"` // stdout, stderr, file
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"` // MB
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"` // days
	Compress   bool   `yaml:"compress"`
}

// SourcesConfig represents the configuration for content sources
type SourcesConfig struct {
	Version int      `yaml:"version"`
	Sources []Source `yaml:"sources"`
}

type WebConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Port        int    `yaml:"port"`
	Host        string `yaml:"host"`
	StaticDir   string `yaml:"static_dir"`
	TemplateDir string `yaml:"template_dir"`
}

type APIConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
	RateLimit
}

// Source represents a single content source configuration
type Source struct {
	ID        string                 `yaml:"id"`
	Name      string                 `yaml:"name"`
	URL       string                 `yaml:"url"`
	Type      string                 `yaml:"type"` // html, json, xml, rss
	Enabled   bool                   `yaml:"enabled"`
	Schedule  string                 `yaml:"schedule"` // cron expression
	Headers   map[string]string      `yaml:"headers"`
	Selectors SourceSelectors        `yaml:"selectors"`
	RateLimit *RateLimit             `yaml:"rate_limit"` // Override global rate limit
	Timeout   *time.Duration         `yaml:"timeout"`    // Override global timeout
	Priority  int                    `yaml:"priority"`   // Higher number = higher priority
	Tags      []string               `yaml:"tags"`
	Metadata  map[string]interface{} `yaml:"metadata"`
}

// SourceSelectors contains CSS/XPath selectors for extracting data
type SourceSelectors struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Content     string   `yaml:"content"`
	Author      string   `yaml:"author"`
	Date        string   `yaml:"date"`
	Categories  []string `yaml:"categories"`
	Tags        []string `yaml:"tags"`
	Images      []string `yaml:"images"`
	Links       []string `yaml:"links"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "content-aggregator",
			Environment: "development",
			Debug:       false,
			Timeout:     30 * time.Second,
			MaxRetries:  3,
		},
		Fetcher: FetcherConfig{
			MaxConcurrentWorkers: 10,
			RequestTimeout:       10 * time.Second,
			MaxIdleConns:         100,
			MaxConnsPerHost:      10,
			UserAgent:            "Content-Aggregator/1.0",
			RespectRobotsTxt:     true,
			RateLimiting: RateLimit{
				Enabled:           true,
				RequestsPerMinute: 5,
				BurstSize:         10,
				PerDomain:         true,
				DefaultDelay:      200 * time.Millisecond,
			},
			RetryPolicy: RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  100 * time.Millisecond,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
				RetryableErrors: []string{
					"timeout",
					"connection refused",
					"temporary failure",
				},
			},
		},
		Parser: ParserConfig{
			MaxConcurrentWorkers: 5,
			BufferSize:           100,
			Timeout:              5 * time.Second,
			Selectors: map[string]interface{}{
				"title":       "title, h1, .title, .headline",
				"description": "meta[name=description], .description, .summary",
				"content":     ".content, .article-body, main, article",
				"author":      ".author, .byline, [rel=author]",
				"date":        ".date, .published, time[datetime]",
			},
		},
		Output: OutputConfig{
			Format:      "json",
			Destination: "stdout",
			Pretty:      true,
			API: APIOutputConfig{
				Enabled:  false,
				Host:     "localhost",
				Port:     8080,
				BasePath: "/api/v1",
			},
			Database: DatabaseConfig{
				Enabled: false,
				Driver:  "sqlite",
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath, sourcesPath string) (*Config, error) {
	// Start with default configuration
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Load sources configuration if path is provided
	if sourcesPath != "" {
		sources, err := LoadSources(sourcesPath)
		if err != nil {
			return nil, fmt.Errorf("error loading sources configuration: %w", err)
		}
		// Merge sources configuration into main config
		config.Sources = *sources
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadSources loads source configurations from a YAML file
func LoadSources(sourcesPath string) (*SourcesConfig, error) {
	// Check if sources file exists
	if _, err := os.Stat(sourcesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sources file not found: %s", sourcesPath)
	}

	// Read sources file
	data, err := os.ReadFile(sourcesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading sources file: %w", err)
	}

	// Parse YAML
	var sources SourcesConfig
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, fmt.Errorf("error parsing sources file: %w", err)
	}

	// Validate sources
	if err := sources.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sources configuration: %w", err)
	}

	return &sources, nil
}

// Validate validates the main configuration
func (c *Config) Validate() error {
	// Validate app config
	if c.App.Name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	if c.App.Timeout <= 0 {
		return fmt.Errorf("app timeout must be positive")
	}

	if c.App.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	// Validate fetcher config
	if c.Fetcher.MaxConcurrentWorkers <= 0 {
		return fmt.Errorf("max concurrent workers must be positive")
	}

	if c.Fetcher.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	// Validate parser config
	if c.Parser.MaxConcurrentWorkers <= 0 {
		return fmt.Errorf("parser max concurrent workers must be positive")
	}

	if c.Parser.BufferSize <= 0 {
		return fmt.Errorf("parser buffer size must be positive")
	}

	// Validate output config
	validFormats := []string{"json", "yaml", "csv", "xml"}
	if !contains(validFormats, c.Output.Format) {
		return fmt.Errorf("invalid output format: %s", c.Output.Format)
	}

	validDestinations := []string{"file", "stdout", "api"}
	if !contains(validDestinations, c.Output.Destination) {
		return fmt.Errorf("invalid output destination: %s", c.Output.Destination)
	}

	if c.Output.Destination == "file" && c.Output.FilePath == "" {
		return fmt.Errorf("file path required when destination is file")
	}

	// Validate logging config
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}

// Validate validates the sources configuration
func (s *SourcesConfig) Validate() error {
	if len(s.Sources) == 0 {
		return fmt.Errorf("at least one source must be defined")
	}

	// Check for duplicate source IDs
	ids := make(map[string]bool)
	for _, source := range s.Sources {
		if source.ID == "" {
			return fmt.Errorf("source ID cannot be empty")
		}

		if ids[source.ID] {
			return fmt.Errorf("duplicate source ID: %s", source.ID)
		}

		ids[source.ID] = true

		if source.URL == "" {
			return fmt.Errorf("source URL cannot be empty for source: %s", source.ID)
		}

		if source.Name == "" {
			return fmt.Errorf("source name cannot be empty for source: %s", source.ID)
		}

		validTypes := []string{"html", "json", "xml", "rss"}
		if !contains(validTypes, source.Type) {
			return fmt.Errorf("invalid source type '%s' for source: %s", source.Type, source.ID)
		}
	}

	return nil
}

// GetEnabledSources returns only enabled sources
func (s *SourcesConfig) GetEnabledSources() []Source {
	var enabled []Source
	for _, source := range s.Sources {
		if source.Enabled {
			enabled = append(enabled, source)
		}
	}
	return enabled
}

// GetSourceByID returns a source by its ID
func (s *SourcesConfig) GetSourceByID(id string) (*Source, bool) {
	for _, source := range s.Sources {
		if source.ID == id {
			return &source, true
		}
	}
	return nil, false
}

// GetSourcesByTag returns sources that have the specified tag
func (s *SourcesConfig) GetSourcesByTag(tag string) []Source {
	var sources []Source
	for _, source := range s.Sources {
		for _, sourceTag := range source.Tags {
			if sourceTag == tag {
				sources = append(sources, source)
				break
			}
		}
	}
	return sources
}

// SaveConfig saves configuration to a YAML file
func (c *Config) SaveConfig(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// SaveSources saves sources configuration to a YAML file
func (s *SourcesConfig) SaveSources(sourcesPath string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("error marshaling sources: %w", err)
	}

	if err := os.WriteFile(sourcesPath, data, 0644); err != nil {
		return fmt.Errorf("error writing sources file: %w", err)
	}

	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
