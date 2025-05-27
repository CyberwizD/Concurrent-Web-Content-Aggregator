package coordinator

import (
	"context"
	"log"
	"sync"
)

// WorkerFunc defines the signature for worker functions
type WorkerFunc func(id int, ctx context.Context)

// WorkerPool manages a pool of workers for concurrent processing
type WorkerPool struct {
	size      int                // Number of workers
	name      string             // Pool name (e.g., "fetcher", "parser")
	wg        sync.WaitGroup     // WaitGroup to track worker lifetimes
	ctx       context.Context    // Context for cancellation & timeout
	cancel    context.CancelFunc // Cancels the context
	isRunning bool               // Tracks if the pool is running
	mu        sync.Mutex         // Mutex to protect isRunning
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
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		log.Printf("Worker pool '%s' is already running", p.name)
		return
	}

	// Create cancellable context
	p.ctx, p.cancel = context.WithCancel(parentCtx)
	p.isRunning = true

	log.Printf("Starting %d workers in pool '%s'", p.size, p.name)

	// Start workers
	p.wg.Add(p.size)

	for i := 0; i < p.size; i++ {
		workerID := i + 1 // Worker IDs start from 1
		go func(id int) {
			defer p.wg.Done()

			log.Printf("Worker %d in pool '%s' started", id, p.name)

			// Run the worker function with the context
			workerFn(id, p.ctx)

			log.Printf("Worker %d in pool '%s' finished", id, p.name)
		}(workerID)
	}
}

// Stop signals all workers to stop and waits for them to finish
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning {
		return
	}

	log.Printf("Stopping worker pool '%s'", p.name)

	// Signal workers to stop
	if p.cancel != nil {
		p.cancel()
	}

	// Wait for workers to finish
	p.wg.Wait()

	p.isRunning = false
	log.Printf("Worker pool '%s' stopped", p.name)
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// Size returns the number of workers in the pool
func (p *WorkerPool) Size() int {
	return p.size
}

// IsRunning returns whether the pool is currently running
func (p *WorkerPool) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.isRunning
}
