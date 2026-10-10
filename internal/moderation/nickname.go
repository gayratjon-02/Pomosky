// Package moderation contains content rules for user-supplied text.
package moderation

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateNickname checks that s is an acceptable viewer nickname.
//
// TODO: add a profanity filter (word list) for the supported languages.
func ValidateNickname(s string) error {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	if n < 2 || n > 24 {
		return errors.New("nickname must be 2-24 characters")
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return errors.New("nickname contains control characters")
		}
	}
	return nil
}
