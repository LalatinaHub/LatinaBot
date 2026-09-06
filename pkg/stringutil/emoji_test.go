package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountryISOToEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Indonesia", "ID", "🇮🇩"},
		{"Singapore", "SG", "🇸🇬"},
		{"United States", "US", "🇺🇸"},
		{"Japan lower", "jp", "🇯🇵"},
		{"Germany with spaces", " de ", "🇩🇪"},
		{"Invalid length 1", "I", "I"},
		{"Invalid length 3", "IDN", "IDN"},
		{"Empty", "", ""},
		{"Contains digit", "1A", "1A"},
		{"Special char", "I#", "I#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountryISOToEmoji(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
