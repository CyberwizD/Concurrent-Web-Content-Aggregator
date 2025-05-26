package coordinator

import (
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/fetcher"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/parser"
)

type WorkerPool struct {
	fetcher *fetcher.Fetcher
	parser  *parser.Parser
}

func NewWorkerPool(maxWorker int, WorkerType string) *WorkerPool {
	var fetcher *fetcher.Fetcher
	var parser *parser.Parser

	switch WorkerType {
	case "fetcher":
		fetcher = fetcher.New()
	case "parser":
		parser = parser.New()
	default:
		return nil, ErrInvalidWorkerType
	}

	return &WorkerPool{
		fetcher: fetcher,
		parser:  parser,
	}
}
