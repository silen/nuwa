package logs

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSetLevel(t *testing.T) {
	t.Parallel()

	originalLevel := Logger().GetLevel()
	t.Cleanup(func() {
		SetLevel(originalLevel)
	})

	SetLevel(logrus.ErrorLevel)
	if Logger().GetLevel() != logrus.ErrorLevel {
		t.Fatalf("expected error level, got %v", Logger().GetLevel())
	}
}

func TestSetOutput(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	originalOutput := Logger().Out
	t.Cleanup(func() {
		SetOutput(originalOutput)
	})

	SetOutput(buf)
	Info("hello")
	if buf.Len() == 0 {
		t.Fatalf("expected log output to be written")
	}
}
