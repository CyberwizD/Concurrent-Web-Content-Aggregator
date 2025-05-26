package parser

import (
	"fmt"
)

type Parser struct {
	HtmlParser *HtmlParser
	JsonParser *JsonParser
	XmlParser  *XmlParser
	RssParser  *RssParser
}

type HtmlParser struct {
	// Add fields and methods specific to HTML parsing
}

type JsonParser struct {
	// Add fields and methods specific to JSON parsing
}

type XmlParser struct {
	// Add fields and methods specific to XML parsing
}

type RssParser struct {
	// Add fields and methods specific to RSS parsing
}

// New creates a new Parser instance with the provided parsers.
func New(htmlParser *HtmlParser, jsonParser *JsonParser, xmlParser *XmlParser, rssParser *RssParser) *Parser {
	return &Parser{
		HtmlParser: htmlParser,
		JsonParser: jsonParser,
		XmlParser:  xmlParser,
		RssParser:  rssParser,
	}
}

// Get returns a parser for the specified type.
func Get(parserType string) (*Parser, error) {
	switch parserType {
	case "html":
		return &Parser{HtmlParser: &HtmlParser{}}, nil
	case "json":
		return &Parser{JsonParser: &JsonParser{}}, nil
	case "xml":
		return &Parser{XmlParser: &XmlParser{}}, nil
	case "rss":
		return &Parser{RssParser: &RssParser{}}, nil
	default:
		return nil, fmt.Errorf("unknown parser type: %s", parserType)
	}
}
