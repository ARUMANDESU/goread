package domain

import (
	"strings"
	"testing"

	v "github.com/ARUMANDESU/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARUMANDESU/goread/backend/pkg/envx"
	vx "github.com/ARUMANDESU/goread/backend/pkg/validationx"
)

const (
	ValidUsername         = "user1"
	ValidPassword         = "examplePass123!@"
	InvalidPasswordShort  = "sh12!@"
	InvalidPasswordFormat = "OnlyABClol"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	type args struct {
		id       UserID
		username string
		password string
	}

	validArgs := func() args {
		return args{NewUserID(), ValidUsername, ValidPassword}
	}

	tests := []struct {
		name        string
		args        func() args
		expectedErr error
	}{
		{
			name: "valid args",
			args: validArgs,
		},
		{
			name: "valid username: exact min length",
			args: func() args {
				a := validArgs()
				a.username = usernameWithLength(MinUserNameLen)
				return a
			},
		},
		{
			name: "valid username: exact max length",
			args: func() args {
				a := validArgs()
				a.username = usernameWithLength(MaxUserNameLen)
				return a
			},
		},
		{
			name: "invalid id: nil",
			args: func() args {
				a := validArgs()
				a.id = UserID{}
				return a
			},
			expectedErr: v.Errors{"id": v.ErrRequired},
		},
		{
			name: "invalid username: blank",
			args: func() args {
				a := validArgs()
				a.username = ""
				return a
			},
			expectedErr: v.Errors{"name": v.ErrRequired},
		},
		{
			name: "invalid username: too short",
			args: func() args {
				a := validArgs()
				a.username = usernameWithLength(MinUserNameLen - 1)
				return a
			},
			expectedErr: v.Errors{"name": v.ErrLengthOutOfRange},
		},
		{
			name: "invalid username: too long",
			args: func() args {
				a := validArgs()
				a.username = usernameWithLength(MaxUserNameLen + 1)
				return a
			},
			expectedErr: v.Errors{"name": v.ErrLengthOutOfRange},
		},
		{
			name: "invalid password: blank",
			args: func() args {
				a := validArgs()
				a.password = ""
				return a
			},
			expectedErr: v.Errors{"password": v.ErrRequired},
		},
		{
			name: "invalid password: too short",
			args: func() args {
				a := validArgs()
				a.password = InvalidPasswordShort
				return a
			},
			expectedErr: v.Errors{"password": vx.ErrInvalidPasswordFormat},
		},
		{
			name: "invalid password: too long",
			args: func() args {
				a := validArgs()
				a.password = passwordWithLength(MaxUserPasswordLen + 1)
				return a
			},
			expectedErr: v.Errors{"password": vx.ErrInvalidPasswordFormat},
		},
		{
			name: "invalid password: format",
			args: func() args {
				a := validArgs()
				a.password = InvalidPasswordFormat
				return a
			},
			expectedErr: v.Errors{"password": vx.ErrInvalidPasswordFormat},
		},
		{
			name: "multiple validation errors",
			args: func() args {
				return args{id: UserID{}, username: "", password: ""}
			},
			expectedErr: v.Errors{
				"id":       v.ErrRequired,
				"password": v.ErrRequired,
				"name":     v.ErrRequired,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := tt.args()
			u, err := NewUser(a.id, a.username, a.password, envx.Test)
			if tt.expectedErr != nil {
				vx.AssertValidationErrors(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, u)
			assert.Equal(t, a.id, u.id)
			assert.Equal(t, a.username, u.name)
			assert.NoError(t, u.ComparePassword(a.password), "password should verify against hash")
		})
	}
}

func usernameWithLength(l int) string {
	return strings.Repeat("a", l)
}

func passwordWithLength(l int) string {
	if l < 4 {
		return strings.Repeat("p", l)
	}
	// Ensure it meets format requirements (letters + digits + special)
	suffix := "1a!"
	return strings.Repeat("P", l-len(suffix)) + suffix
}
