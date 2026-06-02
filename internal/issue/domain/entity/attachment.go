package entity

import (
	"errors"
	"path/filepath"
	"strings"
	"time"
)

const (
	AttachmentEntityIssue   = "issue"
	AttachmentEntityComment = "comment"

	MaxAttachmentSize = 5 * 1024 * 1024

	maxAttachmentIDLength       = 80
	maxAttachmentEntityIDLength = 80
	maxAttachmentFilenameLength = 255
	maxAttachmentPathLength     = 1000
)

type Attachment struct {
	ID          string    `json:"id"`
	EntityType  string    `json:"entityType"`
	EntityID    string    `json:"entityId"`
	Filename    string    `json:"filename"`
	Path        string    `json:"-"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateAttachmentInput struct {
	ID          string
	EntityType  string
	EntityID    string
	Filename    string
	Path        string
	ContentType string
	Size        int64
}

func NormalizeCreateAttachment(input CreateAttachmentInput) (CreateAttachmentInput, error) {
	input.EntityType = strings.TrimSpace(input.EntityType)
	input.EntityID = strings.TrimSpace(input.EntityID)
	input.Filename = filepath.Base(strings.TrimSpace(input.Filename))
	input.Path = strings.TrimSpace(input.Path)
	input.ContentType = strings.TrimSpace(input.ContentType)
	if input.ID == "" {
		return input, errors.New("attachment id is required")
	}
	if runeCount(input.ID) > maxAttachmentIDLength {
		return input, errors.New("attachment id must be 80 characters or fewer")
	}
	if !IsValidAttachmentEntityType(input.EntityType) {
		return input, errors.New("entityType is invalid")
	}
	if input.EntityID == "" {
		return input, errors.New("entityId is required")
	}
	if runeCount(input.EntityID) > maxAttachmentEntityIDLength {
		return input, errors.New("entityId must be 80 characters or fewer")
	}
	if input.Filename == "" || input.Filename == "." {
		return input, errors.New("filename is required")
	}
	if runeCount(input.Filename) > maxAttachmentFilenameLength {
		return input, errors.New("filename must be 255 characters or fewer")
	}
	if input.Path == "" {
		return input, errors.New("path is required")
	}
	if runeCount(input.Path) > maxAttachmentPathLength {
		return input, errors.New("path must be 1000 characters or fewer")
	}
	if !IsAllowedAttachmentContentType(input.ContentType) {
		return input, errors.New("contentType is invalid")
	}
	if input.Size <= 0 {
		return input, errors.New("size is required")
	}
	if input.Size > MaxAttachmentSize {
		return input, errors.New("file must be 5MB or smaller")
	}
	return input, nil
}

func IsValidAttachmentEntityType(value string) bool {
	return value == AttachmentEntityIssue || value == AttachmentEntityComment
}

func IsAllowedAttachmentContentType(value string) bool {
	switch strings.ToLower(value) {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func AttachmentExtension(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
