package domain

import (
	v "github.com/ARUMANDESU/validation"
	"github.com/gofrs/uuid"

	"github.com/ARUMANDESU/goread/backend/pkg/errorx"
)

const (
	MinAuthorNameLen = 2
	MaxAuthorNameLen = 100
)

type AuthorID = uuid.UUID

type Author struct {
	id   AuthorID
	name string
}

func NewAuthorID() AuthorID {
	return uuid.Must(uuid.NewV7())
}

func NewAuthor(
	id AuthorID,
	name string,
) (*Author, error) {
	const op = errorx.Op("domain.NewAuthor")

	err := v.Errors{
		"id":   v.Validate(id, v.Required),
		"name": v.Validate(name, v.Required, v.Length(MinAuthorNameLen, MaxAuthorNameLen)),
	}.Filter()
	if err != nil {
		return nil, op.Wrap(err)
	}

	return &Author{
		id:   id,
		name: name,
	}, nil
}

func (a *Author) ID() AuthorID {
	return a.id
}

func (a *Author) Name() string {
	return a.name
}
