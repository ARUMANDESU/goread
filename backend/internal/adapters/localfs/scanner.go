package localfs

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"

	"github.com/ARUMANDESU/goread/backend/internal/app/sync"
	"github.com/ARUMANDESU/goread/backend/pkg/errorx"
)

type Scanner struct {
	fs fs.FS
}

func NewScanner(path string) Scanner {
	return Scanner{fs: os.DirFS(path)}
}

func (s Scanner) ScanDir(ctx context.Context) (map[sync_app.Path]sync_app.Hash, error) {
	const op = errorx.Op("localfs.Scanner.ScanDir")

	m := make(map[sync_app.Path]sync_app.Hash)
	var errs errorx.Errors

	err := fs.WalkDir(s.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errs.Append(err)
		}
		if d == nil || d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(s.fs, path)
		if err != nil {
			return fmt.Errorf("failed to read file(%s): %w", path, err)
		}

		h := sha256.New()
		_, err = h.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write file(%s) data into hash: %w", path, err)
		}
		m[path] = h.Sum(nil)

		return nil
	})
	if err != nil {
		errs.Append(fmt.Errorf("failed to walk dir: %w", err))
	}

	return m, op.Wrap(errs.Filter())
}
