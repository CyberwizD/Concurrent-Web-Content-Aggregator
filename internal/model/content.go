package model

import (
	"net/http"
	"time"
)

// Content represents the raw fetched content from a web source
// Content represents raw content fetched from a source
type Content struct {
	// Source information
	Source *Source

	// Request information
	URL string

	// Response information
	Body        []byte
	ContentType string
	StatusCode  int
	Headers     http.Header

	// Metadata
	FetchedAt time.Time
}

// Item represents a single piece of content extracted from a source
type Item struct {
	// Core content fields
	ID       string
	Title    string
	Content  string
	URL      string
	Date     time.Time
	Author   string
	Category string

	// Additional fields (varies by source)
	ExtraFields map[string]interface{}

	// Source information
	SourceName string
	SourceURL  string

	// Metadata
	FetchedAt   time.Time
	ParsedAt    time.Time
	ExtractedBy string // Parser that extracted this item
}

// FetchJob represents a job for the fetcher worker pool
type FetchJob struct {
	Source      *Source
	SubmittedAt time.Time
	Metadata    map[string]interface{} // Optional metadata
}

// FetchResult represents the result of a fetch operation
type FetchResult struct {
	Source      *Source
	Content     *Content
	FetchedAt   time.Time
	ProcessedBy int // Worker ID
	Error       error
	Metadata    map[string]interface{} // Optional metadata
}

// ParseJob represents a job for the parser worker pool
type ParseJob struct {
	Source      *Source
	Content     *Content
	SubmittedAt time.Time
	Metadata    map[string]interface{} // Optional metadata
}

// ParseResult represents the result of a parse operation
type ParseResult struct {
	Source      *Source
	Items       []Item
	ParsedAt    time.Time
	ProcessedBy int // Worker ID
	Error       error
	Metadata    map[string]interface{} // Optional metadata
}
