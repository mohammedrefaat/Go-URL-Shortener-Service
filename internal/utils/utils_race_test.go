package utils

import (
	"sync"
	"testing"
)

func TestGenerateID_Concurrent(t *testing.T) {
	// Reset global nodes map to isolate runs.
	nodes = sync.Map{}

	const goroutines = 1000
	const perGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	seen := sync.Map{} // map[string]struct{}

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				id := GenerateID(1)
				if id == "" {
					t.Errorf("empty id generated")
					return
				}
				// Check uniqueness (best-effort)
				if _, loaded := seen.LoadOrStore(id, struct{}{}); loaded {
					// duplicate found — Record an error but do not stop immediately (a very rare coincidence may occur)
					t.Errorf("duplicate id observed: %s", id)
				}
			}
		}()
	}

	wg.Wait()
}
