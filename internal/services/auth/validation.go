package auth

import (
	"strings"
	"unicode"
)

func normalizeUsername(username string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if len(normalized) < 3 || len(normalized) > 30 {
		return "", ErrInvalidUsername
	}

	for _, character := range normalized {
		if character != '_' && !unicode.IsDigit(character) && (character < 'a' || character > 'z') {
			return "", ErrInvalidUsername
		}
	}
	return normalized, nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidPassword
	}
	return nil
}
