package localfs

import (
	"crypto/sha256"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sync_app "github.com/ARUMANDESU/goread/backend/internal/app/sync"
)

func TestScanner_ScanDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fs       fstest.MapFS
		expected map[sync_app.Path]sync_app.Hash
	}{
		{
			name: "multiple file types in nested dirs",
			fs: fstest.MapFS{
				"books/book1/book1.epub":   &fstest.MapFile{Data: []byte("book 1")},
				"books/book2/book2.epub":   &fstest.MapFile{Data: []byte("book 2")},
				"comics/comic1/comic1.cbz": &fstest.MapFile{Data: []byte("comic 1")},
				"manga/manga1/manga.cbr":   &fstest.MapFile{Data: []byte("manga 1")},
			},
			expected: map[sync_app.Path]sync_app.Hash{
				"books/book1/book1.epub":   getDataHash([]byte("book 1")),
				"books/book2/book2.epub":   getDataHash([]byte("book 2")),
				"comics/comic1/comic1.cbz": getDataHash([]byte("comic 1")),
				"manga/manga1/manga.cbr":   getDataHash([]byte("manga 1")),
			},
		},
		{
			name: "single file",
			fs: fstest.MapFS{
				"book.epub": &fstest.MapFile{Data: []byte("data")},
			},
			expected: map[sync_app.Path]sync_app.Hash{
				"book.epub": getDataHash([]byte("data")),
			},
		},
		{
			name:     "empty fs",
			fs:       fstest.MapFS{},
			expected: map[sync_app.Path]sync_app.Hash{},
		},
		{
			name: "deeply nested structure",
			fs: fstest.MapFS{
				"a/b/c/d/deep.epub": &fstest.MapFile{Data: []byte("deep")},
			},
			expected: map[sync_app.Path]sync_app.Hash{
				"a/b/c/d/deep.epub": getDataHash([]byte("deep")),
			},
		},
		{
			name: "files with same content have same hash",
			fs: fstest.MapFS{
				"a.epub": &fstest.MapFile{Data: []byte("same")},
				"b.epub": &fstest.MapFile{Data: []byte("same")},
			},
			expected: map[sync_app.Path]sync_app.Hash{
				"a.epub": getDataHash([]byte("same")),
				"b.epub": getDataHash([]byte("same")),
			},
		},
		{
			name: "files with different content have different hashes",
			fs: fstest.MapFS{
				"a.epub": &fstest.MapFile{Data: []byte("content a")},
				"b.epub": &fstest.MapFile{Data: []byte("content b")},
			},
			expected: map[sync_app.Path]sync_app.Hash{
				"a.epub": getDataHash([]byte("content a")),
				"b.epub": getDataHash([]byte("content b")),
			},
		},
		{
			name: "empty file",
			fs: fstest.MapFS{
				"empty.epub": &fstest.MapFile{Data: []byte{}},
			},
			expected: map[sync_app.Path]sync_app.Hash{
				"empty.epub": getDataHash([]byte{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := Scanner{fs: tt.fs}

			res, err := s.ScanDir(t.Context())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func getDataHash(data []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}
