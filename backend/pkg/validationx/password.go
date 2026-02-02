package vx

import (
	"errors"
	"unicode"

	v "github.com/ARUMANDESU/validation"

	"github.com/ARUMANDESU/goread/backend/pkg/i18nx"
)

var ErrInvalidPasswordFormat = v.NewError(i18nx.ValidationInvalidPasswordFormat, i18nx.ValidationInvalidPasswordFormatMessage)

type PasswordFormatRule struct{}

// Validate validates a password string against the defined rules.
// It checks for minimum length, presence of uppercase, lowercase, digit, and special character.
func (r PasswordFormatRule) Validate(value any) error {
	password, ok := value.(string)
	if !ok {
		return errors.New("value is not a string")
	}

	if len(password) < 8 {
		return ErrInvalidPasswordFormat
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		default:
			return ErrInvalidPasswordFormat
		}
	}

	if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
		return ErrInvalidPasswordFormat
	}

	return nil
}
