package cmd

import (
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"8 characters", 8},
		{"16 characters", 16},
		{"32 characters", 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRandomString(tt.length)

			if len(result) != tt.length {
				t.Errorf("generateRandomString(%d) = %q (length %d), want length %d",
					tt.length, result, len(result), tt.length)
			}

			if result == "" {
				t.Error("generateRandomString() returned empty string")
			}
		})
	}
}

func TestGenerateRandomStringUniqueness(t *testing.T) {
	s1 := generateRandomString(16)
	s2 := generateRandomString(16)

	if s1 == s2 {
		t.Error("generateRandomString() generated duplicate strings")
	}
}
