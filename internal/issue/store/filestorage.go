package store

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type AttachmentStorage struct {
	home string
}

func NewAttachmentStorageFromHome() (*AttachmentStorage, error) {
	home, err := tqconfig.EnsureHome()
	if err != nil {
		return nil, err
	}
	return NewAttachmentStorage(home), nil
}

func NewAttachmentStorage(home string) *AttachmentStorage {
	return &AttachmentStorage{home: home}
}

func (s *AttachmentStorage) Save(entityType string, entityID string, filename string, contentType string, data []byte) (entity.CreateAttachmentInput, error) {
	id, err := randomAttachmentID()
	if err != nil {
		return entity.CreateAttachmentInput{}, err
	}
	ext := entity.AttachmentExtension(contentType)
	relativePath := filepath.Join("system", "data", "attachments", entityType, entityID, id+ext)
	absolutePath, err := s.Resolve(relativePath)
	if err != nil {
		return entity.CreateAttachmentInput{}, err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return entity.CreateAttachmentInput{}, fmt.Errorf("create attachment dir: %w", err)
	}
	if err := os.WriteFile(absolutePath, data, 0o644); err != nil {
		return entity.CreateAttachmentInput{}, fmt.Errorf("write attachment file: %w", err)
	}
	return entity.CreateAttachmentInput{
		ID:          id,
		EntityType:  entityType,
		EntityID:    entityID,
		Filename:    filename,
		Path:        relativePath,
		ContentType: contentType,
		Size:        int64(len(data)),
	}, nil
}

func (s *AttachmentStorage) Resolve(relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("attachment path must be relative")
	}
	clean := filepath.Clean(relativePath)
	if clean == "." || clean == ".." || len(clean) >= 3 && clean[:3] == "../" {
		return "", fmt.Errorf("attachment path escapes home")
	}
	full := filepath.Join(s.home, clean)
	rel, err := filepath.Rel(s.home, full)
	if err != nil {
		return "", err
	}
	if rel == ".." || len(rel) >= 3 && rel[:3] == "../" {
		return "", fmt.Errorf("attachment path escapes home")
	}
	return full, nil
}

func (s *AttachmentStorage) Open(relativePath string) (*os.File, error) {
	path, err := s.Resolve(relativePath)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *AttachmentStorage) Delete(relativePath string) error {
	path, err := s.Resolve(relativePath)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ReadAttachmentBytes(reader io.Reader, maxSize int64) ([]byte, error) {
	var buf bytes.Buffer
	limit := maxSize + 1
	if _, err := io.CopyN(&buf, reader, limit); err != nil && err != io.EOF {
		return nil, err
	}
	if int64(buf.Len()) > maxSize {
		return nil, fmt.Errorf("file must be 5MB or smaller")
	}
	return buf.Bytes(), nil
}

func randomAttachmentID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate attachment id: %w", err)
	}
	return "att_" + hex.EncodeToString(raw[:]), nil
}
