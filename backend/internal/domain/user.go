package domain

import (
	v "github.com/ARUMANDESU/validation"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ARUMANDESU/goread/backend/pkg/envx"
	"github.com/ARUMANDESU/goread/backend/pkg/errorx"
	vx "github.com/ARUMANDESU/goread/backend/pkg/validationx"
)

const (
	PasswordCostFactor = 12
	MinUserNameLen     = 3
	MaxUserNameLen     = 75
	MinUserPasswordLen = 8
	MaxUserPasswordLen = 128
)

const (
	ValidUsername         = "user1"
	ValidPassword         = "examplePass123!@"
	InvalidPasswordShort  = "sh12!@"
	InvalidPasswordFormat = "OnlyABClol"
)

type UserID = uuid.UUID

//go:generate go tool gobuildergen --type User
type User struct {
	id       UserID `builder:"default=NewUserID()"`
	name     string `builder:"default=ValidUsername"`
	passHash []byte
}

func NewUserID() UserID {
	return uuid.Must(uuid.NewV7())
}

func NewUser(
	id UserID,
	name string,
	password string,
	mode envx.Mode,
) (*User, error) {
	const op = errorx.Op("domain.NewUser")

	err := v.Errors{
		"id":       v.Validate(id, vx.Required),
		"name":     v.Validate(name, vx.Required, v.Length(MinUserNameLen, MaxUserNameLen)),
		"password": v.Validate(password, vx.Required, vx.Password(MinUserPasswordLen, MaxUserPasswordLen)),
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
