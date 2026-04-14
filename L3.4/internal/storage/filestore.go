package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// FileStore отвечает за сохранение файлов на диск.
type FileStore struct {
	originalDir  string
	processedDir string
	thumbsDir    string
}

// NewFileStore создает файловое хранилище.
func NewFileStore(originalDir, processedDir, thumbsDir string) *FileStore {
	return &FileStore{
		originalDir:  originalDir,
		processedDir: processedDir,
		thumbsDir:    thumbsDir,
	}
}

// EnsureDirs создает папки, если они отсутствуют.
func (s *FileStore) EnsureDirs() error {
	dirs := []string{s.originalDir, s.processedDir, s.thumbsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// SaveOriginal сохраняет исходное изображение.
func (s *FileStore) SaveOriginal(filename string, file multipart.File) (string, error) {
	path := filepath.Join(s.originalDir, filename)

	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return path, nil
}

// DeleteFile удаляет файл, если он существует.
func (s *FileStore) DeleteFile(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ProcessedDir возвращает директорию обработанных файлов.
func (s *FileStore) ProcessedDir() string {
	return s.processedDir
}

// ThumbsDir возвращает директорию миниатюр.
func (s *FileStore) ThumbsDir() string {
	return s.thumbsDir
}
