package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareSnapshots(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		old      LibrarySnapshot
		curr     LibrarySnapshot
		expected CompareSnapshotsResult
	}{
		{
			name:     "both empty",
			old:      LibrarySnapshot{},
			curr:     LibrarySnapshot{},
			expected: CompareSnapshotsResult{},
		},
		{
			name: "no changes",
			old:  LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2")},
			curr: LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2")},
			expected: CompareSnapshotsResult{},
		},
		{
			name: "file added",
			old:  LibrarySnapshot{},
			curr: LibrarySnapshot{"a.epub": []byte("h1")},
			expected: CompareSnapshotsResult{
				Added: []FileInfo{{Hash: []byte("h1"), Path: "a.epub"}},
			},
		},
		{
			name: "multiple files added",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2"), "c.epub": []byte("h3")},
			expected: CompareSnapshotsResult{
				Added: []FileInfo{
					{Hash: []byte("h2"), Path: "b.epub"},
					{Hash: []byte("h3"), Path: "c.epub"},
				},
			},
		},
		{
			name: "file removed",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{},
			expected: CompareSnapshotsResult{
				Removed: []FileInfo{{Hash: []byte("h1"), Path: "a.epub"}},
			},
		},
		{
			name: "multiple files removed",
			old:  LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2"), "c.epub": []byte("h3")},
			curr: LibrarySnapshot{"a.epub": []byte("h1")},
			expected: CompareSnapshotsResult{
				Removed: []FileInfo{
					{Hash: []byte("h2"), Path: "b.epub"},
					{Hash: []byte("h3"), Path: "c.epub"},
				},
			},
		},
		{
			name: "file moved",
			old:  LibrarySnapshot{"old/a.epub": []byte("h1")},
			curr: LibrarySnapshot{"new/a.epub": []byte("h1")},
			expected: CompareSnapshotsResult{
				Moved: []FileInfo{{Hash: []byte("h1"), Path: "old/a.epub", NewPath: "new/a.epub"}},
			},
		},
		{
			name: "file added and file removed",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{"b.epub": []byte("h2")},
			expected: CompareSnapshotsResult{
				Added:   []FileInfo{{Hash: []byte("h2"), Path: "b.epub"}},
				Removed: []FileInfo{{Hash: []byte("h1"), Path: "a.epub"}},
			},
		},
		{
			name: "mixed: added, removed, and moved",
			old:  LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2"), "c.epub": []byte("h3")},
			curr: LibrarySnapshot{"a.epub": []byte("h1"), "d.epub": []byte("h2"), "e.epub": []byte("h4")},
			expected: CompareSnapshotsResult{
				Added:   []FileInfo{{Hash: []byte("h4"), Path: "e.epub"}},
				Removed: []FileInfo{{Hash: []byte("h3"), Path: "c.epub"}},
				Moved:   []FileInfo{{Hash: []byte("h2"), Path: "b.epub", NewPath: "d.epub"}},
			},
		},
		{
			name: "content changed at same path",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{"a.epub": []byte("h2")},
			expected: CompareSnapshotsResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CompareSnapshots(tt.old, tt.curr)

			assert.ElementsMatch(t, tt.expected.Added, result.Added, "Added mismatch")
			assert.ElementsMatch(t, tt.expected.Removed, result.Removed, "Removed mismatch")
			assert.ElementsMatch(t, tt.expected.Moved, result.Moved, "Moved mismatch")
		})
	}
}
