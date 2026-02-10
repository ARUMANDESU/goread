package localfs

import (
	"crypto/sha256"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/ARUMANDESU/goread/backend/internal/domain/value-object"
)

func TestScanner_ScanDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fs       fstest.MapFS
		expected vo.LibrarySnapshot
	}{
		{
			name: "multiple file types in nested dirs",
			fs: fstest.MapFS{
				"books/book1/book1.epub":   &fstest.MapFile{Data: []byte("book 1")},
				"books/book2/book2.epub":   &fstest.MapFile{Data: []byte("book 2")},
				"comics/comic1/comic1.cbz": &fstest.MapFile{Data: []byte("comic 1")},
				"manga/manga1/manga.cbr":   &fstest.MapFile{Data: []byte("manga 1")},
			},
			expected: vo.LibrarySnapshot{
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
			expected: vo.LibrarySnapshot{
				"book.epub": getDataHash([]byte("data")),
			},
		},
		{
			name:     "empty fs",
			fs:       fstest.MapFS{},
			expected: vo.LibrarySnapshot{},
		},
		{
			name: "deeply nested structure",
			fs: fstest.MapFS{
				"a/b/c/d/deep.epub": &fstest.MapFile{Data: []byte("deep")},
			},
			expected: vo.LibrarySnapshot{
				"a/b/c/d/deep.epub": getDataHash([]byte("deep")),
			},
		},
		{
			name: "files with same content have same hash",
			fs: fstest.MapFS{
				"a.epub": &fstest.MapFile{Data: []byte("same")},
				"b.epub": &fstest.MapFile{Data: []byte("same")},
			},
			expected: vo.LibrarySnapshot{
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
			expected: vo.LibrarySnapshot{
				"a.epub": getDataHash([]byte("content a")),
				"b.epub": getDataHash([]byte("content b")),
			},
		},
		{
			name: "empty file",
			fs: fstest.MapFS{
				"empty.epub": &fstest.MapFile{Data: []byte{}},
			},
			expected: vo.LibrarySnapshot{
				"empty.epub": getDataHash([]byte{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := Scanner{fs: tt.fs}

			res, err := s.Snapshot(t.Context())
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
