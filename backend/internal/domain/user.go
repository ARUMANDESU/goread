package domain

import (
	v "github.com/ARUMANDESU/validation"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ARUMANDESU/goread/backend/pkg/envx"
	"github.com/ARUMANDESU/goread/backend/pkg/errorx"
)

const (
	PasswordCostFactor = 12
	MinUserNameLen     = 3
	MaxUserNameLen     = 75
)

type UserID uuid.UUID

type User struct {
	id       UserID
	name     string
	passHash []byte
}

func NewUserID() UserID {
	return UserID(uuid.Must(uuid.NewV7()))
}

func NewUser(
	id UserID,
	name string,
	password string,
	mode envx.Mode,
) (*User, error) {
	const op = errorx.Op("domain.NewUser")

	err := v.Errors{
		"id":       v.Validate(id, v.Required),
		"name":     v.Validate(name, v.Required, v.Length(MinUserNameLen, MaxUserNameLen)),
		"password": v.Validate(password, v.Required, v.Length(8, 128)),
	}.Filter()
	if err != nil {
		return nil, op.Wrap(err)
	}

	passHash, err := NewPasswordHash(password, mode)
	if err != nil {
		return nil, op.Wrap(err)
	}

	return &User{
		id:       id,
		name:     name,
		passHash: passHash,
	}, nil
}

func (u *User) ComparePassword(password string) error {
	const op = errorx.Op("domain.User.ComparePassword")
	return op.Wrap(bcrypt.CompareHashAndPassword(u.passHash, []byte(password)))
}

func NewPasswordHash(password string, mode envx.Mode) ([]byte, error) {
	const op = errorx.Op("domain.NewPasswordHash")
	costFactor := PasswordCostFactor
	if mode == envx.Test {
		costFactor = bcrypt.MinCost
	}
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), costFactor)
	if err != nil {
		return nil, op.WrapMsg(err, "failed to generate password hash")
	}
	return passhash, nil
}
