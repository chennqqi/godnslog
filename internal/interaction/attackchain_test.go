package interaction

import (
	"testing"
)

func TestHigherConfidence(t *testing.T) {
	order := []string{"high", "medium", "low"}
	tests := []struct {
		a, b, expected string
	}{
		{"high", "medium", "high"},
		{"medium", "low", "medium"},
		{"low", "medium", "medium"},
		{"high", "high", "high"},
	}
	for _, tt := range tests {
		got := higherConfidence(tt.a, tt.b, order)
		if got != tt.expected {
			t.Errorf("higherConfidence(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	set := map[string]bool{"dns": true, "http": true, "ldap": true}
	keys := sortedKeys(set)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Errorf("keys not sorted: %v", keys)
			break
		}
	}
}

func TestSortedKeysEmpty(t *testing.T) {
	keys := sortedKeys(map[string]bool{})
	if keys == nil || len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}
