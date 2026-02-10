package vo

import "github.com/ARUMANDESU/goread/backend/pkg/errorx"

type (
	Path = string
	Hash = []byte
)

type LibrarySnapshot map[Path]Hash

type FileInfo struct {
	Hash    []byte
	Path    string
	NewPath string
}

type CompareSnapshotsResult struct {
	Added   []FileInfo
	Removed []FileInfo
	Moved   []FileInfo
}

func CompareSnapshots(old, curr LibrarySnapshot) (CompareSnapshotsResult, error) {
	const op = errorx.Op("vo.CompareSnapshots")

	return CompareSnapshotsResult{}, nil
}
