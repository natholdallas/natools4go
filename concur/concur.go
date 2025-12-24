// Package concur provides primitives for managing concurrent execution of tasks.
package concur

import "sync"

// Go starts all provided tasks in their own goroutines and returns a WaitGroup.
// It is useful when you want to trigger tasks and perform other work
// before waiting for their completion.
func Go(tasks ...func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	for _, task := range tasks {
		// Capture task in a local variable to avoid the closure bug
		t := task
		go func() {
			defer wg.Done()
			// Recommended: Add recover() here to handle potential panics
			defer func() { recover() }()
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
