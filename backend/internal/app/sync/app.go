package sync_app

import (
	"context"

	vo "github.com/ARUMANDESU/goread/backend/internal/domain/value-object"
	"github.com/ARUMANDESU/goread/backend/pkg/errorx"
)

/*
how it will look like in the fs
will leave it here for now
Books
├── Books
│    └── Author
│           ├── Book1
│           │    └── Book1.epub
│           ├── Book2
│           │    └── Book2.epub
│           ├── Book3
│           │    └── Book3.epub
│           ├── Book3
── Book3.epub
│           └── Book4
│                 └── Book4.pdf
└── Comics
│     ├── Plastic Man (1944)
│     │    └── Plastic Man #002 (1944).cbz
│     └── Comic (2008)
│           └── Comic #001 (2008).cbr
└── Manga
      └── Attack on Titan (2012)
          └── Attack on Titan #001 (2012).pdf
*/

// how can I identify items if they got their names changed?
// I should generate ID and save it alongside with library items, in a metadata file?
// this solution sound ok, and I don't have any other useful ideas
//
// upd: we can just hash files to identify them

// scan library -> check if something changed or not
// check for change -> no  --> do nothing
//                  -> yes --> removed
//                         +-> added
//                         +-> mixed

// may be I should save library structure in tree form, then check current and old(from db)
//
// for example:
// old:
// library - books - author1 - book1 - book1.epub
//                           - book2 - book2.epub
//         - comics - Plastic Man (1944) - Plastic Man #002 (1944).cbz
//                  - Comic (2008) - Comic #001 (2008).cbr
//         - manga - Attack on Titan (2012) - Attack on Titan #001 (2012).pdf

// current from fs:
// library - books - author1 - book1 - book1.epub
//                           - book2 - book2.epub
//                           - book3 - book3.epub
//         - comics - Plastic Man (1944) - Plastic Man #002 (1944).cbz
//         - manga - Attack on Titan (2012) - Attack on Titan #001 (2012).pdf
//                 - Chainsawman - Chainsawman.cbr
//
// if we compare them we can see that we have:
// new: books.book3 and manga.Chainsawman
// removed: comics.Comic
//
// but this is too complex may we should we Merkle tree?
// or just map[path]hash

type (
	Path = string
	Hash = []byte
)

type Scanner interface {
	ScanDir(context.Context) (map[Path]Hash, error)
}

type LibraryRepo interface {
	GetLibrary(context.Context) vo.Library
}

type App struct {
	Scanner     Scanner
	LibraryRepo LibraryRepo
}

func (a *App) ScanLibrary(ctx context.Context) error {
	const op = errorx.Op("sync.App.ScanLibrary")
	_, _ = a.Scanner.ScanDir(ctx)
	return nil
}
