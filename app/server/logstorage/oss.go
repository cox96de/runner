package logstorage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/cox96de/runner/util"

	"github.com/cockroachdb/errors"
)

const ErrNotFound = util.StringError("not found")

type Reader interface {
	io.Reader
	io.Seeker
}
type OSS interface {
	Open(ctx context.Context, filename string) (io.ReadCloser, error)
	Save(ctx context.Context, filename string, r Reader) error
}

type FilesystemOSS struct {
	baseDir string
}

func NewFilesystemOSS(baseDir string) *FilesystemOSS {
	return &FilesystemOSS{baseDir: baseDir}
}

func (o *FilesystemOSS) Open(ctx context.Context, filename string) (io.ReadCloser, error) {
	file, err := os.Open(filepath.Join(o.baseDir, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, errors.WithMessage(err, "failed to open file")
	}
	return file, err
}

func (o *FilesystemOSS) Save(ctx context.Context, filename string, r Reader) error {
	fp := filepath.Join(o.baseDir, filename)
	err := os.MkdirAll(filepath.Dir(fp), 0o755)
	if err != nil {
		return errors.WithMessage(err, "failed to create directory")
	}
	file, err := os.Create(fp)
	if err != nil {
		return errors.WithMessage(err, "failed to create file")
	}
	defer file.Close()
	_, err = io.Copy(file, r)
	return err
}
