package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

type Storage interface {
	SaveFile(subDir string, filename string, reader io.Reader) (string, int64, error)
	GetFile(path string) (*os.File, error)
	DeleteFile(path string) error
	AppendChunk(evidenceID uuid.UUID, seq int, data []byte) (string, error)
	AssembleChunks(evidenceID uuid.UUID, totalChunks int, destFilename string) (string, int64, error)
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

func (s *LocalStorage) SaveFile(subDir string, filename string, reader io.Reader) (string, int64, error) {
	dir := filepath.Join(s.baseDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, err
	}

	fullPath := filepath.Join(dir, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	n, err := io.Copy(dst, reader)
	if err != nil {
		return "", 0, err
	}

	return fullPath, n, nil
}

func (s *LocalStorage) GetFile(path string) (*os.File, error) {
	return os.Open(path)
}

func (s *LocalStorage) DeleteFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func (s *LocalStorage) AppendChunk(evidenceID uuid.UUID, seq int, data []byte) (string, error) {
	chunkDir := filepath.Join(s.baseDir, "chunks", evidenceID.String())
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return "", err
	}

	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%05d.bin", seq))
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return "", err
	}

	return chunkPath, nil
}

func (s *LocalStorage) AssembleChunks(evidenceID uuid.UUID, totalChunks int, destFilename string) (string, int64, error) {
	chunkDir := filepath.Join(s.baseDir, "chunks", evidenceID.String())
	outputDir := filepath.Join(s.baseDir, "audio_evidences")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", 0, err
	}

	entries, err := os.ReadDir(chunkDir)
	if err != nil || len(entries) == 0 {
		return "", 0, fmt.Errorf("tidak ada potongan audio yang tersimpan di server")
	}

	var binFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".bin") {
			binFiles = append(binFiles, entry.Name())
		}
	}

	if len(binFiles) == 0 {
		return "", 0, fmt.Errorf("tidak ada potongan audio .bin yang valid di server")
	}

	sort.Strings(binFiles)

	finalPath := filepath.Join(outputDir, destFilename)
	out, err := os.Create(finalPath)
	if err != nil {
		return "", 0, err
	}
	defer out.Close()

	var totalBytes int64 = 0
	for _, fname := range binFiles {
		cPath := filepath.Join(chunkDir, fname)
		cData, err := os.ReadFile(cPath)
		if err != nil {
			continue
		}
		n, err := out.Write(cData)
		if err != nil {
			return "", 0, err
		}
		totalBytes += int64(n)
	}

	// Clean up chunks
	_ = os.RemoveAll(chunkDir)

	return finalPath, totalBytes, nil
}

