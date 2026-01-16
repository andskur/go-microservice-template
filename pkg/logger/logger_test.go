package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogSingleton(t *testing.T) {
	first := Log()
	second := Log()

	if first == nil || second == nil {
		t.Fatalf("expected logger instances, got nil")
	}

	if first != second {
		t.Fatalf("expected singleton instance, got different pointers")
	}
}

func TestLogFormatter(t *testing.T) {
	log := Log()
	formatter, ok := log.Formatter.(*logrus.TextFormatter)
	if !ok {
		t.Fatalf("expected TextFormatter, got %T", log.Formatter)
	}
	if !formatter.FullTimestamp {
		t.Fatalf("expected FullTimestamp to be true")
	}
}
