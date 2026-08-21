package concur

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestGoAndRun(t *testing.T) {
	var n atomic.Int32
	tasks := []func(){func() { n.Add(1) }, func() { n.Add(1) }}
	Run(tasks...)
	if got := n.Load(); got != 2 {
		t.Errorf("n = %d, want 2", got)
	}
}

func TestPoolBoundsConcurrency(t *testing.T) {
	pool := NewPool(2)
	var max, cur int32
	var done atomic.Int32
	const total = 50
	for range total {
		pool.Submit(func() {
			c := atomic.AddInt32(&cur, 1)
			for {
				m := atomic.LoadInt32(&max)
				if c <= m || atomic.CompareAndSwapInt32(&max, m, c) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&cur, -1)
			done.Add(1)
		})
	}
	pool.Close()
	pool.Wait()
	if done.Load() != total {
		t.Errorf("done = %d, want %d", done.Load(), total)
	}
	if m := atomic.LoadInt32(&max); m > 2 {
		t.Errorf("max concurrency = %d, want <= 2", m)
	}
}

func TestPoolSubmitAfterClosePanics(t *testing.T) {
	pool := NewPool(1)
	pool.Close()
	defer func() {
		if recover() == nil {
			t.Error("expected panic on Submit after Close")
		}
	}()
	pool.Submit(func() {})
}
