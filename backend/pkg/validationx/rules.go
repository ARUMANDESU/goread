package vx

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unicode"

	v "github.com/ARUMANDESU/validation"

	"github.com/ARUMANDESU/goread/backend/pkg/i18nx"
)

var ErrInvalidPasswordFormat = v.NewError(i18nx.ValidationInvalidPasswordFormat, i18nx.ValidationInvalidPasswordFormatMessage)

var (
	IsPasswordFromat = PasswordFormatRule{}
	Required         = RequiredRule{}
)

var valuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

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

type RequiredRule struct{}

func (r RequiredRule) Validate(value any) error {
	value, isNil := Indirect(value)
	if isNil || IsEmpty(value) {
		return v.ErrRequired
	}

	return nil
}

// IsEmpty checks if a value is empty or not.
// A value is considered empty if
// - integer, float: zero
// - bool: false
// - string: len() == 0
// - slice, map, array: nil or len() == 0
// - interface, pointer: nil or the referenced value is empty
func IsEmpty(value any) bool {
	v := reflect.ValueOf(value)

	fmt.Printf("%+v; %+v\n", v, v.Kind())
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return v.Equal(reflect.Zero(v.Type())) || v.Len() == 0
	case reflect.String:
		return v.Len() == 0 || v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	case reflect.Struct:
		v, ok := value.(time.Time)
		if ok && v.IsZero() {
			return true
		}
	}

	return false
}

// Indirect returns the value that the given interface or pointer references to.
// If the value implements driver.Valuer, it will deal with the value returned by
// the Value() method instead. A boolean value is also returned to indicate if
// the value is nil or not (only applicable to interface, pointer, map, and slice).
// If the value is neither an interface nor a pointer, it will be returned back.
//
// From: https://github.com/ARUMANDESU/validation/blob/main/util.go#L134
func Indirect(value interface{}) (interface{}, bool) {
	rv := reflect.ValueOf(value)
	kind := rv.Kind()
	switch kind {
	case reflect.Invalid:
		return nil, true
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil, true
		}
		return Indirect(rv.Elem().Interface())
	case reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		if rv.IsNil() {
			return nil, true
		}
	}

	return value, false
}
