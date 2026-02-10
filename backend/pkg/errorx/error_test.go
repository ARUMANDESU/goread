package errorx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		errs     []error
		wantNil  bool
		expected string
	}{
		{
			name:     "multiple non-nil errors",
			errs:     []error{fmt.Errorf("something"), fmt.Errorf("something: x"), fmt.Errorf("lol: kek")},
			expected: "something, something: x, lol: kek",
		},
		{
			name:     "single error",
			errs:     []error{fmt.Errorf("only one")},
			expected: "only one",
		},
		{
			name:    "empty slice",
			errs:    []error{},
			wantNil: true,
		},
		{
			name:    "all nils",
			errs:    []error{nil, nil, nil},
			wantNil: true,
		},
		{
			name:    "single nil",
			errs:    []error{nil},
			wantNil: true,
		},
		{
			name:     "nil in the middle",
			errs:     []error{fmt.Errorf("first"), nil, fmt.Errorf("third")},
			expected: "first, third",
		},
		{
			name:     "nil at the start",
			errs:     []error{nil, fmt.Errorf("second"), fmt.Errorf("third")},
			expected: "second, third",
		},
		{
			name:     "nil at the end",
			errs:     []error{fmt.Errorf("first"), fmt.Errorf("second"), nil},
			expected: "first, second",
		},
		{
			name:     "consecutive nils in the middle",
			errs:     []error{fmt.Errorf("first"), nil, nil, fmt.Errorf("last")},
			expected: "first, last",
		},
		{
			name:     "consecutive nils at the start",
			errs:     []error{nil, nil, fmt.Errorf("third")},
			expected: "third",
		},
		{
			name:     "nils scattered",
			errs:     []error{nil, fmt.Errorf("a"), nil, fmt.Errorf("b"), nil},
			expected: "a, b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := Errors{}

			for _, err := range tt.errs {
				errs.Append(err)
			}

			result := errs.Filter()
			if tt.wantNil {
				assert.Nil(t, result)
				return
			}
			assert.Equal(t, tt.expected, result.Error())
		})
	}
}
