package vx

import (
	"errors"
	"testing"

	v "github.com/ARUMANDESU/validation"
)

// AssertValidationErrors asserts that the error is a map of validation errors (v.Errors)
// and matches the structure and contents of the expected errors.
func AssertValidationErrors(t *testing.T, err error, expected error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error %v, got nil", expected)
	}

	var actualVerrs v.Errors
	if !errors.As(err, &actualVerrs) {
		t.Fatalf("expected error to be of type v.Errors, got %T: %v", err, err)
	}

	var expectedVerrs v.Errors
	if !errors.As(expected, &expectedVerrs) {
		t.Fatalf("expected 'expected' error to be of type v.Errors, got %T: %v", expected, expected)
	}

	for field, actualErr := range actualVerrs {
		if _, found := expectedVerrs[field]; !found {
			t.Errorf("found unexpected validation error on field %q: %v", field, actualErr)
		}
	}

	for field, expectedErr := range expectedVerrs {
		actualErr, found := actualVerrs[field]
		if !found {
			t.Errorf("missing expected validation error on field %q. Expected: %v", field, expectedErr)
			continue
		}

		AssertValidationError(t, actualErr, expectedErr)
	}
}

// AssertValidationError asserts that a specific error matches the expected error
// code and message. It handles both leaf errors (v.Error) and nested maps (v.Errors).
func AssertValidationError(t *testing.T, err error, expected error) {
	t.Helper()

	if err == nil {
		t.Errorf("expected validation error %v, got nil", expected)
		return
	}

	var expectedMap v.Errors
	if errors.As(expected, &expectedMap) {
		AssertValidationErrors(t, err, expected)
		return
	}

	var expectedLeaf v.Error
	if !errors.As(expected, &expectedLeaf) {
		// If expected isn't a v.Error or v.Errors, fall back to standard equality
		if err.Error() != expected.Error() {
			t.Errorf("expected standard error %q, got %q", expected.Error(), err.Error())
		}
		return
	}

	var actualLeaf v.Error
	if !errors.As(err, &actualLeaf) {
		t.Errorf("expected error to be of type v.Error, got %T: %v", err, err)
		return
	}

	if actualLeaf.Code() != expectedLeaf.Code() {
		t.Errorf("error code mismatch.\nExpected: %s\nGot:      %s", expectedLeaf.Code(), actualLeaf.Code())
	}
	if actualLeaf.Message() != expectedLeaf.Message() {
		t.Errorf("error message mismatch.\nExpected: %s\nGot:      %s", expectedLeaf.Message(), actualLeaf.Message())
	}
}
