package task

import (
	"sync"
	"testing"
	"time"
)

func TestSetMaxConcurrent(t *testing.T) {
	r := NewRunner(nil, nil, nil, 2)

	var mu sync.Mutex
	active := 0
	maxActive := 0

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.acquireSlot()
			defer r.releaseSlot()

			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			active--
			mu.Unlock()
		}()
	}

	time.Sleep(20 * time.Millisecond)
	// Increase concurrency while tasks are running.
	r.SetMaxConcurrent(4)

	wg.Wait()

	if maxActive != 4 {
		t.Fatalf("max concurrent active = %d, want 4 (after increase)", maxActive)
	}
}

func TestSetMaxConcurrentDecrease(t *testing.T) {
	r := NewRunner(nil, nil, nil, 4)

	var mu sync.Mutex
	active := 0
	maxActive := 0

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.acquireSlot()
			defer r.releaseSlot()

			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			mu.Unlock()

			time.Sleep(30 * time.Millisecond)

			mu.Lock()
			active--
			mu.Unlock()
		}()
	}

	time.Sleep(10 * time.Millisecond)
	r.SetMaxConcurrent(2)

	wg.Wait()

	if maxActive > 4 {
		t.Fatalf("max concurrent active = %d, want <= 4 (initial limit)", maxActive)
	}
}
