package coordinator

import (
	"context"
	"sync"
)

// WorkerFunc defines the signature for worker functions
type WorkerFunc func(id int, ctx context.Context)

// WorkerPool manages a pool of workers for concurrent processing
type WorkerPool struct {
	size      int
	name      string
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.Mutex
}

// NewWorkerPool creates a new worker pool with the specified size
func NewWorkerPool(size int, name string) *WorkerPool {
	if size <= 0 {
		size = 1 // Ensure at least one worker
	}

	return &WorkerPool{
		size: size,
		name: name,
	}
}

// Start initializes the worker pool and start all workers
func (p *WorkerPool) Start(parentCtx context.Context, workerFn WorkerFunc) {

}
