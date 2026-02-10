package sync_app

import (
	"context"
	"log/slog"

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

type Snapshotter interface {
	Snapshot(context.Context) (vo.LibrarySnapshot, error)
}

type SnapshotRepo interface {
	GetLibrarySnapshot(context.Context) (vo.LibrarySnapshot, error)
}

type App struct {
	Snapshotter  Snapshotter
	SnapshotRepo SnapshotRepo
}

func (a *App) ScanLibrary(ctx context.Context) error {
	const op = errorx.Op("sync.App.ScanLibrary")
	snapshot, err := a.Snapshotter.Snapshot(ctx)
	if err != nil && len(snapshot) == 0 {
		return op.Wrap(err)
	} else if err != nil {
		slog.ErrorContext(ctx, op.Wrap(err).Error())
	}

	oldSnapshot, err := a.SnapshotRepo.GetLibrarySnapshot(ctx)
	if err != nil {
		return op.Wrap(err)
	}

	// TODO compare old and current snapshots to get diffs: added, removed, moved(removed+added)
	_, _ = vo.CompareSnapshots(oldSnapshot, snapshot)

	// get metadata for for added library items

	// get authors

	// save new library items and authors
	// update library items to deleted which where removed
	// update library items' file path for moved ones
	return nil
}
