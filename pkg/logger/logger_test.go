package logger

import "testing"

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
