package cli

import (
	"strings"
	"unicode"
)

func split(s string) []string { return strings.Split(s, ",") }
func Safe(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
			return '�'
		}
		return r
	}, s)
}
