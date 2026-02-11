package vo

type (
	Path    = string
	Hash    = []byte
	HashStr = string
)

type LibrarySnapshot map[Path]Hash

type FileInfo struct {
	Hash    []byte
	Path    string
	NewPath string
}

type CompareSnapshotsResult struct {
	Added   map[HashStr]FileInfo
	Removed map[HashStr]FileInfo
	Moved   map[HashStr]FileInfo
}

func CompareSnapshots(old, curr LibrarySnapshot) CompareSnapshotsResult {
	added := make(map[HashStr]FileInfo)
	moved := make(map[HashStr]FileInfo)
	removed := make(map[HashStr]FileInfo)

	for path, oldHash := range old {
		if _, ok := curr[path]; !ok {
			removed[string(oldHash)] = FileInfo{Hash: oldHash, Path: path}
		}
	}

	for path, currHash := range curr {
		if _, ok := old[path]; !ok {
			if file, ok := removed[string(currHash)]; ok {
				file.NewPath = path
				moved[string(currHash)] = file
				delete(removed, string(currHash))
				continue
			}
			added[string(currHash)] = FileInfo{Hash: currHash, Path: path}
		}
	}

	return CompareSnapshotsResult{
		Added:   added,
		Removed: removed,
		Moved:   moved,
	}
}
