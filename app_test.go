package nuwa

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/silen/nuwa/config"
)

func TestRunEReturnsConfigError(t *testing.T) {
	originalErr := config.ConfigError()
	t.Cleanup(func() {
		config.SetConfigErrorForTest(originalErr)
	})

	wantErr := errors.New("config failed")
	config.SetConfigErrorForTest(wantErr)

	err := RunE()
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped config error, got %v", err)
	}
}

func TestRunContextReturnsConfigError(t *testing.T) {
	originalErr := config.ConfigError()
	t.Cleanup(func() {
		config.SetConfigErrorForTest(originalErr)
	})

	wantErr := errors.New("config failed")
	config.SetConfigErrorForTest(wantErr)

	err := RunContext(context.Background())
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped config error, got %v", err)
	}
}

func TestRunContextReturnsCanceledContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RunContext(ctx)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled context error, got %v", err)
	}
}

func TestNewResetsRoutePrintFilter(t *testing.T) {
	filterRepeatMu.Lock()
	filterRepeat["get_/demo"] = true
	filterRepeatMu.Unlock()

	t.Setenv(environmentKey, "test")
	engine := New()
	if engine == nil {
		t.Fatalf("expected engine to be created")
	}

	filterRepeatMu.Lock()
	defer filterRepeatMu.Unlock()
	if len(filterRepeat) != 0 {
		t.Fatalf("expected filterRepeat to be reset")
	}
}

func TestDefaultReturnsSharedEngine(t *testing.T) {
	originalEngine := Default()
	t.Cleanup(func() {
		SetDefaultEngine(originalEngine)
	})

	engine := gin.New()
	SetDefaultEngine(engine)

	if got := Default(); got != engine {
		t.Fatalf("expected shared engine to be replaced")
	}
}
