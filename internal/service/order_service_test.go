package service

import (
	"testing"
)

func TestGeneratePIN(t *testing.T) {
	seen := make(map[string]bool)

	for range 1000 {
		pin, err := generatePIN()
		if err != nil {
			t.Fatalf("generatePIN() error: %v", err)
		}
		if len(pin) != 6 {
			t.Errorf("expected 6-digit PIN, got %q (len=%d)", pin, len(pin))
		}
		for _, ch := range pin {
			if ch < '0' || ch > '9' {
				t.Errorf("PIN %q contains non-digit character %q", pin, ch)
			}
		}
		seen[pin] = true
	}

	// With 1000 samples from a 10^6 space, expect at least 950 unique values.
	if len(seen) < 950 {
		t.Errorf("PIN entropy too low: only %d unique values in 1000 samples", len(seen))
	}
}
