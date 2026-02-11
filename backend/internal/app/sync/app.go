package sync_app

import (
	"context"
	"log/slog"

	"github.com/ARUMANDESU/goread/backend/internal/domain"
	vo "github.com/ARUMANDESU/goread/backend/internal/domain/value-object"
	"github.com/ARUMANDESU/goread/backend/pkg/dbx"
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

type MetadataExtractor interface {
	Extract(context.Context, []vo.Path) (map[vo.Path]vo.Metadata, error)
}

type SnapshotRepo interface {
	GetLibrarySnapshot(context.Context) (vo.LibrarySnapshot, error)
	ReplaceSnapshot(context.Context, vo.LibrarySnapshot) error
}

type AuthorRepo interface {
	GetOrCreateAuthors(ctx context.Context, names []string) ([]domain.Author, error)
}

type LibraryItemRepo interface {
	CreateLibraryItems(context.Context, []*domain.LibraryItem) error
	GetLibraryItemsByHash(context.Context, []vo.Hash) ([]*domain.LibraryItem, error)
	UpdateLibraryItems(context.Context, []*domain.LibraryItem) error
}

type App struct {
	Session           dbx.Session
	Snapshotter       Snapshotter
	MetadataExtractor MetadataExtractor
	SnapshotRepo      SnapshotRepo
	LibraryItemRepo   LibraryItemRepo
	AuthorRepo        AuthorRepo
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

	results := vo.CompareSnapshots(oldSnapshot, snapshot)

	addedPaths := make([]string, 0, len(results.Added))
	for _, v := range results.Added {
		addedPaths = append(addedPaths, v.Path)
	}

	metadataMap, err := a.MetadataExtractor.Extract(ctx, addedPaths)
	if err != nil {
		return op.Wrap(err)
	}

	uniqueNames := make(map[string]struct{})
	for _, md := range metadataMap {
		for _, name := range md.Authors {
			uniqueNames[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(uniqueNames))
	for name := range uniqueNames {
		names = append(names, name)
	}

	return a.Session.Transaction(ctx, func(ctx context.Context) error {
		authors, err := a.AuthorRepo.GetOrCreateAuthors(ctx, names)
		if err != nil {
			return op.Wrap(err)
		}

		authorByName := make(map[string]domain.AuthorID, len(authors))
		for _, author := range authors {
			authorByName[author.Name()] = author.ID()
		}

		libraryItems := make([]*domain.LibraryItem, 0, len(metadataMap))
		for path, md := range metadataMap {
			ids := make([]domain.AuthorID, len(md.Authors))
			for i, name := range md.Authors {
				ids[i] = authorByName[name]
			}

			item, err := domain.NewLibraryItem(
				domain.NewLibraryItemID(),
				md.Title,
				domain.Book, // TODO parse properly
				ids,
				md.Subjects,
				md.Languages,
				md.Description,
				path,
				snapshot[path],
			)
			if err != nil {
				slog.WarnContext(ctx, "skipping item", "path", path, "error", err)
				delete(snapshot, path)
				continue
			}
			libraryItems = append(libraryItems, item)
		}

		err = a.LibraryItemRepo.CreateLibraryItems(ctx, libraryItems)
		if err != nil {
			return op.Wrap(err)
		}

		hashes := make([]vo.Hash, 0, len(results.Moved)+len(results.Removed))
		for _, v := range results.Moved {
			hashes = append(hashes, v.Hash)
		}
		for _, v := range results.Removed {
			hashes = append(hashes, v.Hash)
		}

		libraryItems, err = a.LibraryItemRepo.GetLibraryItemsByHash(ctx, hashes)
		if err != nil {
			return op.Wrap(err)
		}

		for _, item := range libraryItems {
			if _, ok := results.Removed[string(item.Hash())]; ok {
				item.Delete()
			}
			if f, ok := results.Moved[string(item.Hash())]; ok {
				item.UpdatePath(f.NewPath)
			}
		}

		err = a.LibraryItemRepo.UpdateLibraryItems(ctx, libraryItems)
		if err != nil {
			return op.Wrap(err)
		}

		err = a.SnapshotRepo.ReplaceSnapshot(ctx, snapshot)
		if err != nil {
			return op.Wrap(err)
		}

		return nil
	})
}
