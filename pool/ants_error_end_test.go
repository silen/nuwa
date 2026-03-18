package pool

import (
	"context"
	"errors"
	"testing"
)

func TestExecTaskByGoroutineErrorEndReturnsTaskError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	err := ExecTaskByGoroutineErrorEnd(context.Background(), []int{1, 2, 3}, func(v int) error {
		if v == 2 {
			return wantErr
		}
		return nil
	}, WithPoolSize(2))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected task failure error, got %v", err)
	}
}

func TestExecTaskByGoroutineErrorEndHandlesEmptyParams(t *testing.T) {
	t.Parallel()

	if err := ExecTaskByGoroutineErrorEnd[int](context.Background(), nil, func(v int) error { return nil }); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestExecTaskByGoroutineErrorEndFallsBackToDefaultSizeForNonPositivePool(t *testing.T) {
	t.Parallel()

	if err := ExecTaskByGoroutineErrorEnd(context.Background(), []int{1}, func(v int) error { return nil }, WithPoolSize(0)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestExecTaskByGoroutineErrorEndHandlesNilContext(t *testing.T) {
	t.Parallel()

	if err := ExecTaskByGoroutineErrorEnd[int](nil, []int{1}, func(v int) error { return nil }); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
