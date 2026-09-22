package common

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentRateLimiterInitializationAndLimit(t *testing.T) {
	var limiter InMemoryRateLimiter
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limiter.Init(0)
			if limiter.Request("same-client", 20, 3600) {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 20 {
		t.Fatalf("accepted %d requests, want 20", accepted.Load())
	}
}
