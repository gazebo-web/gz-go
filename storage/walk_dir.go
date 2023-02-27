package storage

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// WalkDirFunc defines the signature of a WalkDirFunc that will be executed on every file when
// walking through a directory.
type WalkDirFunc func(ctx context.Context, path string, body io.Reader) error

// WalkDir finds all the files and directories inside src and executes walkFunc on every file.
func WalkDir(ctx context.Context, src string, walkFunc WalkDirFunc) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return processFile(ctx, path, src, walkFunc)
	})
}

// processFile executes walkFunc in the file found in path.
func processFile(ctx context.Context, path string, src string, walkFunc WalkDirFunc) error {
	key, err := filepath.Rel(src, path)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	err = walkFunc(ctx, key, f)
	if err != nil {
		return err
	}
	return nil
}
