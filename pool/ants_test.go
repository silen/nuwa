package pool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
)

func TestGoFallsBackWhenPoolIsOverloaded(t *testing.T) {
	originalPool := globalPool
	defer func() {
		globalPool = originalPool
	}()

	pool, err := ants.NewPool(1, ants.WithNonblocking(true))
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	defer pool.Release()
	globalPool = pool

	blocked := make(chan struct{})
	if err := globalPool.Submit(func() {
		<-blocked
	}); err != nil {
		t.Fatalf("submit blocker: %v", err)
	}

	ran := make(chan struct{}, 1)
	Go(context.Background(), func() {
		ran <- struct{}{}
	})

	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatalf("expected fallback goroutine to run")
	}

	close(blocked)
}

func TestExecTaskByGoroutineHandlesEmptyParams(t *testing.T) {
	t.Parallel()

	if err := ExecTaskByGoroutine[int](context.Background(), nil, func(v int) error { return nil }, WithPoolSize(1)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestExecTaskByGoroutineFallsBackToDefaultSizeForNonPositivePool(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var got []int
	if err := ExecTaskByGoroutine(context.Background(), []int{1, 2}, func(v int) error {
		mu.Lock()
		got = append(got, v)
		mu.Unlock()
		return nil
	}, WithPoolSize(0)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected all tasks to run, got %d", len(got))
	}
}

func TestExecTaskByGoroutineHandlesNilContext(t *testing.T) {
	t.Parallel()

	if err := ExecTaskByGoroutine[int](nil, []int{1}, func(v int) error { return nil }); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
