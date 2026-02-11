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
			expected: CompareSnapshotsResult{Added: map[HashStr]FileInfo{}, Removed: map[HashStr]FileInfo{}, Moved: map[HashStr]FileInfo{}},
		},
		{
			name:     "no changes",
			old:      LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2")},
			curr:     LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2")},
			expected: CompareSnapshotsResult{Added: map[HashStr]FileInfo{}, Removed: map[HashStr]FileInfo{}, Moved: map[HashStr]FileInfo{}},
		},
		{
			name: "file added",
			old:  LibrarySnapshot{},
			curr: LibrarySnapshot{"a.epub": []byte("h1")},
			expected: CompareSnapshotsResult{
				Added:   map[HashStr]FileInfo{"h1": {Hash: []byte("h1"), Path: "a.epub"}},
				Removed: map[HashStr]FileInfo{},
				Moved:   map[HashStr]FileInfo{},
			},
		},
		{
			name: "multiple files added",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2"), "c.epub": []byte("h3")},
			expected: CompareSnapshotsResult{
				Added: map[HashStr]FileInfo{
					"h2": {Hash: []byte("h2"), Path: "b.epub"},
					"h3": {Hash: []byte("h3"), Path: "c.epub"},
				},
				Removed: map[HashStr]FileInfo{},
				Moved:   map[HashStr]FileInfo{},
			},
		},
		{
			name: "file removed",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{},
			expected: CompareSnapshotsResult{
				Added:   map[HashStr]FileInfo{},
				Removed: map[HashStr]FileInfo{"h1": {Hash: []byte("h1"), Path: "a.epub"}},
				Moved:   map[HashStr]FileInfo{},
			},
		},
		{
			name: "multiple files removed",
			old:  LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2"), "c.epub": []byte("h3")},
			curr: LibrarySnapshot{"a.epub": []byte("h1")},
			expected: CompareSnapshotsResult{
				Added: map[HashStr]FileInfo{},
				Removed: map[HashStr]FileInfo{
					"h2": {Hash: []byte("h2"), Path: "b.epub"},
					"h3": {Hash: []byte("h3"), Path: "c.epub"},
				},
				Moved: map[HashStr]FileInfo{},
			},
		},
		{
			name: "file moved",
			old:  LibrarySnapshot{"old/a.epub": []byte("h1")},
			curr: LibrarySnapshot{"new/a.epub": []byte("h1")},
			expected: CompareSnapshotsResult{
				Added:   map[HashStr]FileInfo{},
				Removed: map[HashStr]FileInfo{},
				Moved:   map[HashStr]FileInfo{"h1": {Hash: []byte("h1"), Path: "old/a.epub", NewPath: "new/a.epub"}},
			},
		},
		{
			name: "file added and file removed",
			old:  LibrarySnapshot{"a.epub": []byte("h1")},
			curr: LibrarySnapshot{"b.epub": []byte("h2")},
			expected: CompareSnapshotsResult{
				Added:   map[HashStr]FileInfo{"h2": {Hash: []byte("h2"), Path: "b.epub"}},
				Removed: map[HashStr]FileInfo{"h1": {Hash: []byte("h1"), Path: "a.epub"}},
				Moved:   map[HashStr]FileInfo{},
			},
		},
		{
			name: "mixed: added, removed, and moved",
			old:  LibrarySnapshot{"a.epub": []byte("h1"), "b.epub": []byte("h2"), "c.epub": []byte("h3")},
			curr: LibrarySnapshot{"a.epub": []byte("h1"), "d.epub": []byte("h2"), "e.epub": []byte("h4")},
			expected: CompareSnapshotsResult{
				Added:   map[HashStr]FileInfo{"h4": {Hash: []byte("h4"), Path: "e.epub"}},
				Removed: map[HashStr]FileInfo{"h3": {Hash: []byte("h3"), Path: "c.epub"}},
				Moved:   map[HashStr]FileInfo{"h2": {Hash: []byte("h2"), Path: "b.epub", NewPath: "d.epub"}},
			},
		},
		{
			name:     "content changed at same path",
			old:      LibrarySnapshot{"a.epub": []byte("h1")},
			curr:     LibrarySnapshot{"a.epub": []byte("h2")},
			expected: CompareSnapshotsResult{Added: map[HashStr]FileInfo{}, Removed: map[HashStr]FileInfo{}, Moved: map[HashStr]FileInfo{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CompareSnapshots(tt.old, tt.curr)

			assert.Equal(t, tt.expected.Added, result.Added, "Added mismatch")
			assert.Equal(t, tt.expected.Removed, result.Removed, "Removed mismatch")
			assert.Equal(t, tt.expected.Moved, result.Moved, "Moved mismatch")
		})
	}
}
