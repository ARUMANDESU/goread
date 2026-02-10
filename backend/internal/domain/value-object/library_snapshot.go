package vo

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

func CompareSnapshots(old, curr LibrarySnapshot) CompareSnapshotsResult {
	var added []FileInfo
	var moved []FileInfo
	removedMap := make(map[string]FileInfo)

	for path, oldHash := range old {
		if _, ok := curr[path]; !ok {
			removedMap[string(oldHash)] = FileInfo{Hash: oldHash, Path: path}
		}
	}

	for path, currHash := range curr {
		if _, ok := old[path]; !ok {
			if file, ok := removedMap[string(currHash)]; ok {
				file.NewPath = path
				moved = append(moved, file)
				delete(removedMap, string(currHash))
				continue
			}
			added = append(added, FileInfo{Hash: currHash, Path: path})
		}
	}

	removed := make([]FileInfo, 0, len(removedMap))
	for _, file := range removedMap {
		removed = append(removed, file)
	}

	return CompareSnapshotsResult{
		Added:   added,
		Removed: removed,
		Moved:   moved,
	}
}
