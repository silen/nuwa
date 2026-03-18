package token

import (
	"context"
	"errors"
	"testing"
)

func TestCheckEmptyTokenReturnsNotFound(t *testing.T) {
	t.Parallel()

	_, err := NewToken(context.Background()).Check("")
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}
}
