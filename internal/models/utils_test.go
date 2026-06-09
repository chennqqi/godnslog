package models

import (
	"testing"
)

func TestComputeDeterministicHash(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
		validate bool
	}{
		{
			name: "simple map",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr:  false,
			validate: true,
		},
		{
			name: "nested map",
			data: map[string]interface{}{
				"key1": "value1",
				"nested": map[string]interface{}{
					"nested_key": "nested_value",
				},
			},
			wantErr:  false,
			validate: true,
		},
		{
			name: "map with array",
			data: map[string]interface{}{
				"items": []string{"item1", "item2", "item3"},
				"count": 3,
			},
			wantErr:  false,
			validate: true,
		},
		{
			name: "same data produces same hash",
			data: map[string]interface{}{
				"z_key": "value",
				"a_key": "value",
			},
			wantErr:  false,
			validate: true,
		},
		{
			name: "different key order produces same hash",
			data: map[string]interface{}{
				"a_key": "value",
				"z_key": "value",
			},
			wantErr:  false,
			validate: true,
		},
		{
			name: "nil data",
			data: nil,
			wantErr: false,
			validate: true,
		},
	}

	hashes := make(map[string]string)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ComputeDeterministicHash(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeDeterministicHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate {
				if len(hash) != 64 {
					t.Errorf("ComputeDeterministicHash() returned hash length = %d, want 64", len(hash))
				}

				// Store hash for comparison tests
				hashes[tt.name] = hash
			}
		})
	}

	// Verify that same data (regardless of key order) produces same hash
	if hashes["same data produces same hash"] != hashes["different key order produces same hash"] {
		t.Errorf("Hashes for same data with different key order should match: %s != %s",
			hashes["same data produces same hash"],
			hashes["different key order produces same hash"])
	}
}
