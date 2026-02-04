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

	tests := []struct {
		name        string
		id          UserID
		userName    string
		password    string
		expectedErr error
	}{
		{
			name:     "valid args",
			id:       NewUserID(),
			password: ValidPassword,
			userName: ValidUsername,
		},
		{
			name:        "invalid id: nil",
			id:          UserID{},
			password:    ValidPassword,
			userName:    ValidUsername,
			expectedErr: v.Errors{"id": v.ErrRequired},
		},
		{
			name:        "invalid username: too short",
			id:          NewUserID(),
			password:    ValidPassword,
			userName:    usernameWithLength(MinUserNameLen - 1),
			expectedErr: v.Errors{"name": v.ErrLengthOutOfRange},
		},
		{
			name:        "invalid username: too long",
			id:          NewUserID(),
			password:    ValidPassword,
			userName:    usernameWithLength(MaxUserNameLen + 1),
			expectedErr: v.Errors{"name": v.ErrLengthOutOfRange},
		},
		{
			name:        "invalid password: empty",
			id:          NewUserID(),
			password:    "",
			userName:    ValidUsername,
			expectedErr: v.Errors{"password": v.ErrRequired},
		},
		{
			name:        "invalid password: too short",
			id:          NewUserID(),
			password:    InvalidPasswordShort,
			userName:    ValidUsername,
			expectedErr: v.Errors{"password": v.ErrLengthOutOfRange},
		},
		{
			name:        "invalid password: too long",
			id:          NewUserID(),
			password:    passwordWithLength(129),
			userName:    ValidUsername,
			expectedErr: v.Errors{"password": v.ErrLengthOutOfRange},
		},
		{
			name:        "invalid password: format",
			id:          NewUserID(),
			password:    InvalidPasswordFormat,
			userName:    ValidUsername,
			expectedErr: v.Errors{"password": vx.ErrInvalidPasswordFormat},
		},
		{
			name:     "multiple validation errors",
			id:       UserID{},
			password: "",
			userName: "",
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

			u, err := NewUser(tt.id, tt.userName, tt.password, envx.Test)
			if tt.expectedErr != nil {
				vx.AssertValidationErrors(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, u)
			assert.Equal(t, tt.id, u.id)
			assert.Equal(t, tt.userName, u.name)
			assert.NoError(t, u.ComparePassword(tt.password), "password should verify against hash")
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
