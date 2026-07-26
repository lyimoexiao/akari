package textureadapter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileStorage implements skinlib.FileStorage using the local filesystem.
type FileStorage struct {
	dir string
}

func NewFileStorage(dir string) *FileStorage {
	return &FileStorage{dir: dir}
}

func (s *FileStorage) Save(ctx context.Context, reader io.Reader) (hash string, size int64, err error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", 0, fmt.Errorf("create storage dir: %w", err)
	}

	h := sha256.New()
	tee := io.TeeReader(reader, h)

	tmpFile, err := os.CreateTemp(s.dir, "upload-*")
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	written, err := io.Copy(tmpFile, tee)
	if err != nil {
		tmpFile.Close()
		return "", 0, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	hash = fmt.Sprintf("%x", h.Sum(nil))
	size = written

	// Rename to final hash path (first 2 chars as subdirectory)
	destPath := s.filePath(hash)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return "", 0, fmt.Errorf("create subdirectory: %w", err)
	}
	if err := os.Rename(tmpFile.Name(), destPath); err != nil {
		// If destination already exists (duplicate), that's fine
		if _, statErr := os.Stat(destPath); statErr == nil {
			return hash, size, nil
		}
		return "", 0, fmt.Errorf("move file: %w", err)
	}

	return hash, size, nil
}

func (s *FileStorage) Open(ctx context.Context, hash string) (io.ReadCloser, error) {
	f, err := os.Open(s.filePath(hash))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("texture file not found: %s", hash)
		}
		return nil, fmt.Errorf("open texture file: %w", err)
	}
	return f, nil
}

func (s *FileStorage) Delete(ctx context.Context, hash string) error {
	if err := os.Remove(s.filePath(hash)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete texture file: %w", err)
	}
	return nil
}

func (s *FileStorage) Exists(ctx context.Context, hash string) (bool, error) {
	_, err := os.Stat(s.filePath(hash))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("check texture file: %w", err)
}

func (s *FileStorage) filePath(hash string) string {
	if len(hash) < 2 {
		return filepath.Join(s.dir, hash)
	}
	return filepath.Join(s.dir, hash[:2], hash)
}
