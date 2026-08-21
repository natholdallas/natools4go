// Package concur provides primitives for managing concurrent execution of tasks.
package concur

import (
	"runtime"
	"sync"
)

// panicHandler is invoked whenever a task panics. It defaults to re-panicking
// after recording, so the caller can decide. Override via SetPanicHandler.
var (
	panicHandlerMu sync.RWMutex
	panicHandler   func(v any, stack []byte) = nil
)

// SetPanicHandler overrides how recovered panics are reported. When nil (the
// default), panics are re-raised via runtime.Goexit-free panic after logging
// the stack trace.
func SetPanicHandler(fn func(v any, stack []byte)) {
	panicHandlerMu.Lock()
	panicHandler = fn
	panicHandlerMu.Unlock()
}

func recoverPanic() {
	if r := recover(); r != nil {
		stack := make([]byte, 64<<10)
		n := runtime.Stack(stack, false)
		panicHandlerMu.RLock()
		fn := panicHandler
		panicHandlerMu.RUnlock()
		if fn != nil {
			fn(r, stack[:n])
			return
		}
		panic(r)
	}
}

// Go starts all provided tasks in their own goroutines and returns a WaitGroup.
// It is useful when you want to trigger tasks and perform other work
// before waiting for their completion.
func Go(tasks ...func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	for _, task := range tasks {
		t := task
		go func() {
			defer wg.Done()
			defer recoverPanic()
			t()
		}()
	}
	return &wg
}

// Run executes the given tasks concurrently and blocks until all of them are finished.
// This is the simplest way to execute multiple independent operations in parallel.
func Run(tasks ...func()) {
	Go(tasks...).Wait()
}

// Pool is a bounded worker pool that limits the number of concurrently running
// tasks. It is useful for throttling CPU- or IO-heavy work.
type Pool struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	workers chan struct{}
	closed  bool
	once    sync.Once
}

// NewPool creates a worker pool with at most workers concurrent tasks.
// If workers <= 0 it defaults to runtime.NumCPU().
func NewPool(workers int) *Pool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	return &Pool{workers: make(chan struct{}, workers)}
}

// Submit queues a task for execution. After Close, submissions panic to signal
// misuse. The pool waits for all queued tasks via Wait.
func (p *Pool) Submit(task func()) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		panic("concur: Submit called after Close")
	}
	p.wg.Add(1)
	p.mu.Unlock()

	go func() {
		defer p.wg.Done()
		defer recoverPanic()
		p.workers <- struct{}{}
		defer func() { <-p.workers }()
		task()
	}()
}

// Wait blocks until all submitted tasks have finished.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Close prevents further submissions. It does not wait for running tasks;
// call Wait first if you need to synchronize.
func (p *Pool) Close() {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		p.mu.Unlock()
	})
}
