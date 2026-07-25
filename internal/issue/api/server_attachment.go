package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

func (s *Server) attachments(w http.ResponseWriter, r *http.Request) {
	entityType := strings.TrimSpace(r.URL.Query().Get("entity_type"))
	entityID := strings.TrimSpace(r.URL.Query().Get("entity_id"))
	items, err := s.store.AttachmentsByEntity(r.Context(), entityType, entityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "attachments.list.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createAttachment(w http.ResponseWriter, r *http.Request) {
	storage, ok := s.resolveAttachmentStorage(w, "attachments.create")
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(entity.MaxAttachmentSize + 1024*1024); err != nil {
		writeError(w, http.StatusBadRequest, "attachments.create.invalid_request", err)
		return
	}
	entityType := strings.TrimSpace(r.FormValue("entity_type"))
	entityID := strings.TrimSpace(r.FormValue("entity_id"))
	if err := s.ensureAttachmentParent(r.Context(), entityType, entityID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "attachments.create.entity_not_found", errors.New("entity not found"))
			return
		}
		writeError(w, http.StatusBadRequest, "attachments.create.invalid_input", err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "attachments.create.file_required", errors.New("file is required"))
		return
	}
	defer file.Close()

	data, err := store.ReadAttachmentBytes(file, entity.MaxAttachmentSize)
	if err != nil {
		writeError(w, http.StatusBadRequest, "attachments.create.invalid_file", err)
		return
	}
	contentType := normalizeAttachmentContentType(header.Header.Get("Content-Type"), data)
	if !entity.IsAllowedAttachmentContentType(contentType) {
		writeError(w, http.StatusBadRequest, "attachments.create.invalid_file_type", errors.New("file must be PNG, JPEG, GIF, or WebP"))
		return
	}
	input, err := storage.Save(entityType, entityID, header.Filename, contentType, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "attachments.create.storage_error", err)
		return
	}
	created, err := s.store.CreateAttachment(r.Context(), input)
	if err != nil {
		_ = storage.Delete(input.Path)
		writeError(w, http.StatusBadRequest, "attachments.create.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) attachmentContent(w http.ResponseWriter, r *http.Request) {
	storage, ok := s.resolveAttachmentStorage(w, "attachments.content")
	if !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	item, err := s.store.Attachment(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "attachments.content", "attachment")
		return
	}
	file, err := storage.Open(item.Path)
	if err != nil {
		writeError(w, http.StatusNotFound, "attachments.content.file_not_found", err)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(item.Size, 10))
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": item.Filename}))
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("write attachment content %s: %v", item.ID, err)
	}
}

func (s *Server) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	storage, ok := s.resolveAttachmentStorage(w, "attachments.delete")
	if !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	deleted, err := s.store.DeleteAttachment(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "attachments.delete", "attachment")
		return
	}
	if err := storage.Delete(deleted.Path); err != nil {
		log.Printf("delete attachment file %s: %v", deleted.ID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ensureAttachmentParent(ctx context.Context, entityType string, entityID string) error {
	switch entityType {
	case entity.AttachmentEntityIssue:
		id, err := strconv.ParseInt(entityID, 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("entityId is invalid")
		}
		_, err = s.store.Issue(ctx, id)
		return err
	case entity.AttachmentEntityComment:
		id, err := strconv.ParseInt(entityID, 10, 64)
		if err != nil || id <= 0 {
			return fmt.Errorf("entityId is invalid")
		}
		_, err = s.store.Comment(ctx, id)
		return err
	default:
		return fmt.Errorf("entityType is invalid")
	}
}

func (s *Server) resolveAttachmentStorage(w http.ResponseWriter, action string) (*store.AttachmentStorage, bool) {
	if s.attachmentStorage != nil {
		return s.attachmentStorage, true
	}
	storage, err := store.NewAttachmentStorageFromHome()
	if err != nil {
		writeError(w, http.StatusInternalServerError, action+".storage_unavailable", err)
		return nil, false
	}
	return storage, true
}

func normalizeAttachmentContentType(headerContentType string, data []byte) string {
	headerContentType = strings.ToLower(strings.TrimSpace(strings.Split(headerContentType, ";")[0]))
	detected := strings.ToLower(strings.TrimSpace(http.DetectContentType(data)))
	if entity.IsAllowedAttachmentContentType(detected) {
		return detected
	}
	if entity.IsAllowedAttachmentContentType(headerContentType) {
		return headerContentType
	}
	return detected
}
