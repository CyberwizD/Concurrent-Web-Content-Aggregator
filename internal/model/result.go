package model

import (
	"fmt"
	"time"
)

// ResultType defines the category of information in the result
type ResultType string

const (
	// Result type constants
	ResultTypeArticle  ResultType = "article"
	ResultTypeNews     ResultType = "news"
	ResultTypeBlogPost ResultType = "blog_post"
	ResultTypeProduct  ResultType = "product"
	ResultTypeEvent    ResultType = "event"
	ResultTypeMedia    ResultType = "media"
	ResultTypeOther    ResultType = "other"
)

// ResultItem represents a single piece of extracted and normalized data
type ResultItem struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`   // ID of the source
	ContentID   string                 `json:"content_id"`  // ID of the content it was extracted from
	URL         string                 `json:"url"`         // Original URL
	Title       string                 `json:"title"`       // Title of the content
	Description string                 `json:"description"` // Short description or summary
	Content     string                 `json:"content"`     // Full content text if available
	Type        ResultType             `json:"type"`        // Type of result
	Timestamp   time.Time              `json:"timestamp"`   // Publication time if available
	Author      string                 `json:"author,omitempty"`
	Categories  []string               `json:"categories,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Images      []string               `json:"images,omitempty"`   // URLs of related images
	Links       []string               `json:"links,omitempty"`    // Related links
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // Additional metadata
	CreatedAt   time.Time              `json:"created_at"`         // When this result was created
}

// NewResultItem creates a new ResultItem with the current timestamp
func NewResultItem(sourceID, contentID, url string) *ResultItem {
	now := time.Now()

	return &ResultItem{
		ID:        generateID(),
		SourceID:  sourceID,
		ContentID: contentID,
		URL:       url,
		Type:      ResultTypeOther,
		Timestamp: now,
		CreatedAt: now,
	}
}

// SetTitle sets the title of the result
func (r *ResultItem) SetTitle(title string) {
	r.Title = title
}

// SetDescription sets the description of the result
func (r *ResultItem) SetDescription(description string) {
	r.Description = description
}

// SetContent sets the full content of the result
func (r *ResultItem) SetContent(content string) {
	r.Content = content
}

// SetType sets the type of the result
func (r *ResultItem) SetType(resultType ResultType) {
	if resultType == "" {
		r.Type = ResultTypeOther
	} else {
		r.Type = resultType
	}
}

// SetTimestamp sets the publication timestamp of the result
func (r *ResultItem) SetTimestamp(timestamp time.Time) {
	if !timestamp.IsZero() {
		r.Timestamp = timestamp
	} else {
		r.Timestamp = time.Now()
	}
}

// SetAuthor sets the author of the result
func (r *ResultItem) SetAuthor(author string) {
	r.Author = author
}

// AddCategory adds a category to the result
func (r *ResultItem) AddCategory(category string) {
	if r.Categories == nil {
		r.Categories = make([]string, 0)
	}

	r.Categories = append(r.Categories, category)
}

// AddTag adds a tag to the result
func (r *ResultItem) AddTag(tag string) {
	if r.Tags == nil {
		r.Tags = make([]string, 0)
	}

	r.Tags = append(r.Tags, tag)
}

// AddImage adds an image URL to the result
func (r *ResultItem) AddImage(imageURL string) {
	if r.Images == nil {
		r.Images = make([]string, 0)
	}

	r.Images = append(r.Images, imageURL)
}

// AddLink adds a related link to the result
func (r *ResultItem) AddLink(link string) {
	if r.Links == nil {
		r.Links = make([]string, 0)
	}

	r.Links = append(r.Links, link)
}

// AddMetadata adds a metadata key-value pair to the result
func (r *ResultItem) AddMetadata(key string, value interface{}) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]interface{})
	}

	r.Metadata[key] = value
}

// AggregatedResults represents a collection of result items with statistics
type AggregatedResults struct {
	ID               string       `json:"id"`
	Items            []ResultItem `json:"items"`               // All result items
	SourceCount      int          `json:"source_count"`        // Number of sources processed
	SuccessfulCount  int          `json:"successful_count"`    // Number of successful fetches
	FailedCount      int          `json:"failed_count"`        // Number of failed fetches
	TotalFetchTimeMs int64        `json:"total_fetch_time_ms"` // Total time spent fetching
	CreatedAt        time.Time    `json:"created_at"`          // When aggregation was completed
}

// NewAggregatedResults creates a new aggregated results container
func NewAggregatedResults() *AggregatedResults {
	return &AggregatedResults{
		ID:        generateID(),
		Items:     make([]ResultItem, 0),
		CreatedAt: time.Now(),
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// AddItem adds a result item to the aggregated results
func (ar *AggregatedResults) AddItem(item ResultItem) {
	ar.Items = append(ar.Items, item)
}

// AddItems adds multiple result items to the aggregated results
func (ar *AggregatedResults) AddItems(items []ResultItem) {
	ar.Items = append(ar.Items, items...)
}

// UpdateStats updates the statistics for the aggregated results
func (ar *AggregatedResults) UpdateStats(sourceCount, SuccessfulCount, FailedCount int, totalFetchTimeMs int64) {
	ar.SourceCount = sourceCount
	ar.SuccessfulCount = SuccessfulCount
	ar.FailedCount = FailedCount
	ar.TotalFetchTimeMs = totalFetchTimeMs
}
