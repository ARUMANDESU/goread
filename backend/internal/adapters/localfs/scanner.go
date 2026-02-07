package localfs

import (
	"context"
	"io/fs"
	"os"

	"github.com/ARUMANDESU/goread/backend/internal/app/sync"
)

type Scanner struct {
	fs fs.FS
}

func NewScanner(path string) Scanner {
	return Scanner{fs: os.DirFS(path)}
}

func (s Scanner) ScanDir(ctx context.Context) map[sync_app.Path]sync_app.Hash {
	// TODO implement ScanDir
	return nil
}
