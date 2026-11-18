package concur

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestRunSwallowsPanic is the regression test for the default panic handler:
// re-panicking inside a goroutine cannot be recovered by the caller and takes
// the whole process down. A panicking task must not do that.
func TestRunSwallowsPanic(t *testing.T) {
	ran := atomic.Int32{}
	SetPanicHandler(nil)
	t.Cleanup(func() { SetPanicHandler(nil) })

	// If the default handler re-panicked, this call would kill the test binary.
	Run(
		func() { ran.Add(1) },
		func() { panic("boom") },
		func() { ran.Add(1) },
	)

	if got := ran.Load(); got != 2 {
		t.Errorf("surviving tasks ran = %d, want 2", got)
	}
}

// TestPanicHandlerReceivesPanic proves an explicit handler still observes the
// panic and its stack, so fail-fast callers can opt back in.
func TestPanicHandlerReceivesPanic(t *testing.T) {
	var (
		mu       sync.Mutex
		gotValue any
		gotStack int
	)
	SetPanicHandler(func(v any, stack []byte) {
		mu.Lock()
		defer mu.Unlock()
		gotValue = v
		gotStack = len(stack)
	})
	t.Cleanup(func() { SetPanicHandler(nil) })

	Run(func() { panic("custom") })

	mu.Lock()
	defer mu.Unlock()
	if gotValue != "custom" {
		t.Errorf("handler value = %v, want custom", gotValue)
	}
	if gotStack == 0 {
		t.Error("handler received an empty stack trace")
	}
}

// TestGoReturnsWaitGroupAcrossPanic ensures the WaitGroup is still released so a
// panicking task cannot deadlock the caller.
func TestGoReturnsWaitGroupAcrossPanic(t *testing.T) {
	SetPanicHandler(nil)
	t.Cleanup(func() { SetPanicHandler(nil) })

	wg := Go(func() { panic("boom") })
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	<-done // hangs (test timeout) if WaitGroup.Done is skipped
}
