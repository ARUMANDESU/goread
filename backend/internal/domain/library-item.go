package domain

import (
	v "github.com/ARUMANDESU/validation"
	"github.com/gofrs/uuid"

	"gitlab.com/ARUMANDESU/goread/backend/pkg/errorx"
)

const (
	MinLibraryItemTitleLen = 2
	MaxLibraryItemTitleLen = 150
	MinAnnotationLen       = 2
	MaxAnnotationLen       = 2000
)

type LibraryItemID uuid.UUID

type LibraryItemType string

const (
	Book  LibraryItemType = "book"
	Manga LibraryItemType = "manga"
	Comic LibraryItemType = "comic"
)

type LibraryItem struct {
	id         LibraryItemID
	title      string
	itemType   LibraryItemType
	authorIDs  []AuthorID
	genre      []string
	languages  []string
	annotation string
}

func NewLibraryItemID() LibraryItemID {
	return LibraryItemID(uuid.Must(uuid.NewV7()))
}

func NewLibraryItem(
	id LibraryItemID,
	title string,
	itemType LibraryItemType,
	authorIDs []AuthorID,
	genre []string,
	languages []string,
	annotation string,
) (*LibraryItem, error) {
	const op = errorx.Op("domain.NewLibraryItem")
	err := v.Errors{
		"id":         v.Validate(id, v.Required),
		"title":      v.Validate(title, v.Required, v.Length(MinLibraryItemTitleLen, MaxLibraryItemTitleLen)),
		"itemType":   v.Validate(itemType, v.Required, v.In(Book, Manga, Comic)),
		"authorIDs":  v.Validate(authorIDs, v.Required, v.Length(1, 0)),
		"genre":      v.Validate(genre, v.Length(1, 0)),
		"languages":  v.Validate(languages, v.Length(1, 0)),
		"annotation": v.Validate(annotation, v.Length(MinAnnotationLen, MaxAnnotationLen)),
	}.Filter()
	if err != nil {
		return nil, op.Wrap(err)
	}

	return &LibraryItem{
		id:         id,
		title:      title,
		itemType:   itemType,
		authorIDs:  authorIDs,
		genre:      genre,
		languages:  languages,
		annotation: annotation,
	}, nil
}
