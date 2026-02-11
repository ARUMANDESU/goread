package sync_app

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ARUMANDESU/goread/backend/internal/domain"
	vo "github.com/ARUMANDESU/goread/backend/internal/domain/value-object"
	"github.com/ARUMANDESU/goread/backend/pkg/dbx"
)

// --- Mocks ---

type mockSnapshotter struct{ mock.Mock }

func (m *mockSnapshotter) Snapshot(ctx context.Context) (vo.LibrarySnapshot, error) {
	args := m.Called(ctx)
	s, _ := args.Get(0).(vo.LibrarySnapshot)
	return s, args.Error(1)
}

type mockMetadataExtractor struct{ mock.Mock }

func (m *mockMetadataExtractor) Extract(ctx context.Context, paths []vo.Path) (map[vo.Path]vo.Metadata, error) {
	args := m.Called(ctx, paths)
	r, _ := args.Get(0).(map[vo.Path]vo.Metadata)
	return r, args.Error(1)
}

type mockSnapshotRepo struct{ mock.Mock }

func (m *mockSnapshotRepo) GetLibrarySnapshot(ctx context.Context) (vo.LibrarySnapshot, error) {
	args := m.Called(ctx)
	s, _ := args.Get(0).(vo.LibrarySnapshot)
	return s, args.Error(1)
}

func (m *mockSnapshotRepo) ReplaceSnapshot(ctx context.Context, s vo.LibrarySnapshot) error {
	return m.Called(ctx, s).Error(0)
}

type mockAuthorRepo struct{ mock.Mock }

func (m *mockAuthorRepo) GetOrCreateAuthors(ctx context.Context, names []string) ([]domain.Author, error) {
	args := m.Called(ctx, names)
	a, _ := args.Get(0).([]domain.Author)
	return a, args.Error(1)
}

type mockLibraryItemRepo struct{ mock.Mock }

func (m *mockLibraryItemRepo) CreateLibraryItems(ctx context.Context, items []*domain.LibraryItem) error {
	return m.Called(ctx, items).Error(0)
}

func (m *mockLibraryItemRepo) GetLibraryItemsByHash(ctx context.Context, hashes []vo.Hash) ([]*domain.LibraryItem, error) {
	args := m.Called(ctx, hashes)
	items, _ := args.Get(0).([]*domain.LibraryItem)
	return items, args.Error(1)
}

func (m *mockLibraryItemRepo) UpdateLibraryItems(ctx context.Context, items []*domain.LibraryItem) error {
	return m.Called(ctx, items).Error(0)
}

type mockSession struct{ mock.Mock }

func (m *mockSession) Transaction(ctx context.Context, f func(context.Context) error) error {
	m.Called(ctx, f)
	return f(ctx)
}

func (m *mockSession) Begin(ctx context.Context) (dbx.Session, error) {
	args := m.Called(ctx)
	s, _ := args.Get(0).(dbx.Session)
	return s, args.Error(1)
}

func (m *mockSession) Rollback() error        { return m.Called().Error(0) }
func (m *mockSession) Commit() error           { return m.Called().Error(0) }
func (m *mockSession) Context() context.Context { return m.Called().Get(0).(context.Context) }

// --- Helpers ---

func newTestApp(t *testing.T) (
	*App,
	*mockSnapshotter,
	*mockMetadataExtractor,
	*mockSnapshotRepo,
	*mockAuthorRepo,
	*mockLibraryItemRepo,
	*mockSession,
) {
	t.Helper()
	snap := new(mockSnapshotter)
	ext := new(mockMetadataExtractor)
	sr := new(mockSnapshotRepo)
	ar := new(mockAuthorRepo)
	ir := new(mockLibraryItemRepo)
	sess := new(mockSession)

	app := &App{
		Session:           sess,
		Snapshotter:       snap,
		MetadataExtractor: ext,
		SnapshotRepo:      sr,
		LibraryItemRepo:   ir,
		AuthorRepo:        ar,
	}
	return app, snap, ext, sr, ar, ir, sess
}

func mustNewAuthor(t *testing.T, name string) domain.Author {
	t.Helper()
	a, err := domain.NewAuthor(domain.NewAuthorID(), name)
	require.NoError(t, err)
	return *a
}

func mustNewLibraryItem(t *testing.T, path string, hash []byte, authorIDs []domain.AuthorID) *domain.LibraryItem {
	t.Helper()
	item, err := domain.NewLibraryItem(
		domain.NewLibraryItemID(),
		"Test Title",
		domain.Book,
		authorIDs,
		[]string{"fiction"},
		[]string{"en"},
		"A test book",
		path,
		hash,
	)
	require.NoError(t, err)
	return item
}

func matchStrings(expected ...string) any {
	return mock.MatchedBy(func(actual []string) bool {
		if len(actual) != len(expected) {
			return false
		}
		a := slices.Clone(actual)
		e := slices.Clone(expected)
		slices.Sort(a)
		slices.Sort(e)
		return slices.Equal(a, e)
	})
}

func matchHashes(expected ...vo.Hash) any {
	return mock.MatchedBy(func(actual []vo.Hash) bool {
		if len(actual) != len(expected) {
			return false
		}
		a := make([]string, len(actual))
		for i, h := range actual {
			a[i] = string(h)
		}
		e := make([]string, len(expected))
		for i, h := range expected {
			e[i] = string(h)
		}
		slices.Sort(a)
		slices.Sort(e)
		return slices.Equal(a, e)
	})
}

// validMeta returns a Metadata that will pass NewLibraryItem validation.
func validMeta(title, author string) vo.Metadata {
	return vo.Metadata{
		Title:       title,
		Authors:     []string{author},
		Subjects:    []string{"fiction"},
		Languages:   []string{"en"},
		Description: "A description",
	}
}

// setupFullTx sets up all mocks for a full successful transaction pass-through
// with no added/removed/moved items. Callers can override individual On() calls
// before this if needed (first matching On wins in testify).
func setupEmptyTx(
	ext *mockMetadataExtractor,
	sr *mockSnapshotRepo,
	ar *mockAuthorRepo,
	ir *mockLibraryItemRepo,
	sess *mockSession,
) {
	ext.On("Extract", mock.Anything, mock.Anything).Return(map[vo.Path]vo.Metadata{}, nil)
	sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
	ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return([]domain.Author{}, nil)
	ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
	ir.On("GetLibraryItemsByHash", mock.Anything, mock.Anything).Return([]*domain.LibraryItem{}, nil)
	ir.On("UpdateLibraryItems", mock.Anything, mock.Anything).Return(nil)
	sr.On("ReplaceSnapshot", mock.Anything, mock.Anything).Return(nil)
}

// --- Tests ---

func TestScanLibrary(t *testing.T) {
	t.Parallel()

	t.Run("empty filesystem and empty repo", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		setupEmptyTx(ext, sr, ar, ir, sess)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	t.Run("first scan with new files", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		hash1 := []byte("hash1")
		hash2 := []byte("hash2")
		current := vo.LibrarySnapshot{
			"books/book1.epub": hash1,
			"books/book2.epub": hash2,
		}
		author1 := mustNewAuthor(t, "Author One")
		author2 := mustNewAuthor(t, "Author Two")

		snap.On("Snapshot", mock.Anything).Return(current, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, matchStrings("books/book1.epub", "books/book2.epub")).
			Return(map[vo.Path]vo.Metadata{
				"books/book1.epub": validMeta("Book One", "Author One"),
				"books/book2.epub": validMeta("Book Two", "Author Two"),
			}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, matchStrings("Author One", "Author Two")).
			Return([]domain.Author{author1, author2}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.MatchedBy(func(items []*domain.LibraryItem) bool {
			return len(items) == 2
		})).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, mock.MatchedBy(func(h []vo.Hash) bool {
			return len(h) == 0
		})).Return([]*domain.LibraryItem{}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, current).Return(nil)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	t.Run("no changes between scans", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		current := vo.LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2")}
		old := vo.LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2")}

		snap.On("Snapshot", mock.Anything).Return(current, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(old, nil)
		setupEmptyTx(ext, sr, ar, ir, sess)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	t.Run("only removals", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		hash1 := []byte("h1")
		hash2 := []byte("h2")
		author := mustNewAuthor(t, "Author")
		item1 := mustNewLibraryItem(t, "a.epub", hash1, []domain.AuthorID{author.ID()})
		item2 := mustNewLibraryItem(t, "b.epub", hash2, []domain.AuthorID{author.ID()})

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(
			vo.LibrarySnapshot{"a.epub": hash1, "b.epub": hash2}, nil,
		)
		ext.On("Extract", mock.Anything, mock.Anything).Return(map[vo.Path]vo.Metadata{}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return([]domain.Author{}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, matchHashes(hash1, hash2)).
			Return([]*domain.LibraryItem{item1, item2}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.MatchedBy(func(items []*domain.LibraryItem) bool {
			return len(items) == 2
		})).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, mock.Anything).Return(nil)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	t.Run("only moves", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		hash1 := []byte("h1")
		author := mustNewAuthor(t, "Author")
		item := mustNewLibraryItem(t, "old/a.epub", hash1, []domain.AuthorID{author.ID()})

		current := vo.LibrarySnapshot{"new/a.epub": hash1}

		snap.On("Snapshot", mock.Anything).Return(current, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(
			vo.LibrarySnapshot{"old/a.epub": hash1}, nil,
		)
		ext.On("Extract", mock.Anything, mock.MatchedBy(func(p []string) bool {
			return len(p) == 0
		})).Return(map[vo.Path]vo.Metadata{}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return([]domain.Author{}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, matchHashes(hash1)).
			Return([]*domain.LibraryItem{item}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.MatchedBy(func(items []*domain.LibraryItem) bool {
			return len(items) == 1
		})).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, current).Return(nil)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	t.Run("mixed added removed and moved", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		hashA := []byte("hA") // stays
		hashB := []byte("hB") // moved
		hashC := []byte("hC") // removed
		hashD := []byte("hD") // added

		author := mustNewAuthor(t, "Author")
		movedItem := mustNewLibraryItem(t, "old/b.epub", hashB, []domain.AuthorID{author.ID()})
		removedItem := mustNewLibraryItem(t, "c.epub", hashC, []domain.AuthorID{author.ID()})

		old := vo.LibrarySnapshot{
			"a.epub":     hashA,
			"old/b.epub": hashB,
			"c.epub":     hashC,
		}
		current := vo.LibrarySnapshot{
			"a.epub":     hashA,
			"new/b.epub": hashB,
			"d.epub":     hashD,
		}

		snap.On("Snapshot", mock.Anything).Return(current, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(old, nil)
		ext.On("Extract", mock.Anything, matchStrings("d.epub")).
			Return(map[vo.Path]vo.Metadata{
				"d.epub": validMeta("Book D", "Author"),
			}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, matchStrings("Author")).
			Return([]domain.Author{author}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.MatchedBy(func(items []*domain.LibraryItem) bool {
			return len(items) == 1
		})).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, matchHashes(hashB, hashC)).
			Return([]*domain.LibraryItem{movedItem, removedItem}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.MatchedBy(func(items []*domain.LibraryItem) bool {
			return len(items) == 2
		})).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, current).Return(nil)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	// --- Error propagation ---

	t.Run("snapshotter error with empty snapshot returns error", func(t *testing.T) {
		t.Parallel()
		app, snap, _, _, _, _, _ := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{}, errors.New("scan failed"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "scan failed")
	})

	t.Run("snapshotter partial error continues", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		current := vo.LibrarySnapshot{"a.epub": []byte("h1")}
		author := mustNewAuthor(t, "Author")

		snap.On("Snapshot", mock.Anything).Return(current, errors.New("partial"))
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, mock.Anything).
			Return(map[vo.Path]vo.Metadata{
				"a.epub": validMeta("Book A", "Author"),
			}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).
			Return([]domain.Author{author}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, mock.Anything).Return([]*domain.LibraryItem{}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, mock.Anything).Return(nil)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})

	t.Run("GetLibrarySnapshot error", func(t *testing.T) {
		t.Parallel()
		app, snap, _, sr, _, _, _ := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{"a.epub": []byte("h1")}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(nil, errors.New("db error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "db error")
	})

	t.Run("MetadataExtractor error", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, _, _, _ := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{"a.epub": []byte("h1")}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, mock.Anything).Return(nil, errors.New("extract error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "extract error")
	})

	t.Run("GetOrCreateAuthors error", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, _, sess := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{"a.epub": []byte("h1")}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, mock.Anything).
			Return(map[vo.Path]vo.Metadata{
				"a.epub": validMeta("Book", "Author"),
			}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return(nil, errors.New("author error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "author error")
	})

	t.Run("CreateLibraryItems error", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		author := mustNewAuthor(t, "Author")

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{"a.epub": []byte("h1")}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, mock.Anything).
			Return(map[vo.Path]vo.Metadata{
				"a.epub": validMeta("Book A", "Author"),
			}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).
			Return([]domain.Author{author}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(errors.New("create error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "create error")
	})

	t.Run("GetLibraryItemsByHash error", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(
			vo.LibrarySnapshot{"a.epub": []byte("h1")}, nil,
		)
		ext.On("Extract", mock.Anything, mock.Anything).Return(map[vo.Path]vo.Metadata{}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return([]domain.Author{}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, mock.Anything).
			Return(nil, errors.New("hash lookup error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "hash lookup error")
	})

	t.Run("UpdateLibraryItems error", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		hash1 := []byte("h1")
		author := mustNewAuthor(t, "Author")
		item := mustNewLibraryItem(t, "a.epub", hash1, []domain.AuthorID{author.ID()})

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(
			vo.LibrarySnapshot{"a.epub": hash1}, nil,
		)
		ext.On("Extract", mock.Anything, mock.Anything).Return(map[vo.Path]vo.Metadata{}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return([]domain.Author{}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, mock.Anything).
			Return([]*domain.LibraryItem{item}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.Anything).Return(errors.New("update error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "update error")
	})

	t.Run("ReplaceSnapshot error", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		snap.On("Snapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, mock.Anything).Return(map[vo.Path]vo.Metadata{}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, mock.Anything).Return([]domain.Author{}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, mock.Anything).Return([]*domain.LibraryItem{}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, mock.Anything).Return(errors.New("replace error"))

		err := app.ScanLibrary(context.Background())
		require.ErrorContains(t, err, "replace error")
	})

	// --- Edge case ---

	t.Run("NewLibraryItem validation fails skips item and removes from snapshot", func(t *testing.T) {
		t.Parallel()
		app, snap, ext, sr, ar, ir, sess := newTestApp(t)

		hash1 := []byte("h1")
		hash2 := []byte("h2")
		author := mustNewAuthor(t, "Author")

		current := vo.LibrarySnapshot{
			"good.epub": hash1,
			"bad.epub":  hash2,
		}

		snap.On("Snapshot", mock.Anything).Return(current, nil)
		sr.On("GetLibrarySnapshot", mock.Anything).Return(vo.LibrarySnapshot{}, nil)
		ext.On("Extract", mock.Anything, matchStrings("good.epub", "bad.epub")).
			Return(map[vo.Path]vo.Metadata{
				"good.epub": validMeta("Good Book", "Author"),
				"bad.epub": {
					Title:       "", // empty title triggers validation failure
					Authors:     []string{"Author"},
					Subjects:    []string{"fiction"},
					Languages:   []string{"en"},
					Description: "Bad desc",
				},
			}, nil)
		sess.On("Transaction", mock.Anything, mock.Anything).Return(nil)
		ar.On("GetOrCreateAuthors", mock.Anything, matchStrings("Author")).
			Return([]domain.Author{author}, nil)
		ir.On("CreateLibraryItems", mock.Anything, mock.MatchedBy(func(items []*domain.LibraryItem) bool {
			return len(items) == 1
		})).Return(nil)
		ir.On("GetLibraryItemsByHash", mock.Anything, mock.Anything).Return([]*domain.LibraryItem{}, nil)
		ir.On("UpdateLibraryItems", mock.Anything, mock.Anything).Return(nil)
		sr.On("ReplaceSnapshot", mock.Anything, mock.MatchedBy(func(s vo.LibrarySnapshot) bool {
			_, hasBad := s["bad.epub"]
			_, hasGood := s["good.epub"]
			return !hasBad && hasGood && len(s) == 1
		})).Return(nil)

		err := app.ScanLibrary(context.Background())
		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, snap, ext, sr, ar, ir, sess)
	})
}
