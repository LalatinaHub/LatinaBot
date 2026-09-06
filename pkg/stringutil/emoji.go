package stringutil

import (
	"strings"
	"unicode"
)

// CountryISOToEmoji converts a 2-letter ISO country code into a regional indicator flag emoji.
// Examples: "ID" -> 🇮🇩, "SG" -> 🇸🇬, "US" -> 🇺🇸.
func CountryISOToEmoji(isoCode string) string {
	isoCode = strings.TrimSpace(isoCode)
	if len(isoCode) != 2 {
		return isoCode
	}
	for _, r := range isoCode {
		if unicode.IsDigit(r) || !unicode.IsLetter(r) {
			return isoCode
		}
	}
	var b strings.Builder
	for _, r := range strings.ToUpper(isoCode) {
		b.WriteRune(r + 127397)
	}
	return b.String()
}
