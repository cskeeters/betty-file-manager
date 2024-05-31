package main

import (
	"strings"
	"unicode"
)

func IsLower(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func LowerIf(s string, b bool) string {
	if (b) {
		return strings.ToLower(s)
	}
	return s
}
