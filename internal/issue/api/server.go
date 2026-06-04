package api

import (
	"context"
	"database/sql"
	"encoding/json"
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

type Server struct {
	store             *store.Store
	attachmentStorage *store.AttachmentStorage
}

func NewServer(issueStore *store.Store) *Server {
	return &Server{store: issueStore}
}

func NewServerWithAttachmentStorage(issueStore *store.Store, attachmentStorage *store.AttachmentStorage) *Server {
	return &Server{store: issueStore, attachmentStorage: attachmentStorage}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/summary", s.summary)
	mux.HandleFunc("GET /api/v1/projects", s.projects)
	mux.HandleFunc("POST /api/v1/projects", s.createProject)
	mux.HandleFunc("GET /api/v1/projects/{id}", s.project)
	mux.HandleFunc("PATCH /api/v1/projects/{id}", s.updateProject)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", s.deleteProject)
	mux.HandleFunc("POST /api/v1/projects/{id}/check", s.checkProject)
	mux.HandleFunc("GET /api/v1/issues", s.issues)
	mux.HandleFunc("POST /api/v1/issues", s.createIssue)
	mux.HandleFunc("POST /api/v1/issues/states", s.issueStates)
	mux.HandleFunc("GET /api/v1/issues/{id}", s.issue)
	mux.HandleFunc("PATCH /api/v1/issues/{id}", s.updateIssue)
	mux.HandleFunc("GET /api/v1/issues/{issueId}/comments", s.comments)
	mux.HandleFunc("POST /api/v1/issues/{issueId}/comments", s.createComment)
	mux.HandleFunc("PATCH /api/v1/comments/{id}", s.updateComment)
	mux.HandleFunc("GET /api/v1/attachments", s.attachments)
	mux.HandleFunc("POST /api/v1/attachments", s.createAttachment)
	mux.HandleFunc("GET /api/v1/attachments/{id}/content", s.attachmentContent)
	mux.HandleFunc("DELETE /api/v1/attachments/{id}", s.deleteAttachment)
	return withCORS(withLogging(mux))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.store.Summary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "summary.get.internal_error", err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) projects(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Projects(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "projects.list.internal_error", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var input entity.CreateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "projects.create.invalid_request", err)
		return
	}
	created, err := s.store.CreateProject(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "projects.create.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) project(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.get")
	if !ok {
		return
	}
	item, err := s.store.Project(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "projects.get", "project")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.update")
	if !ok {
		return
	}
	var input entity.UpdateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "projects.update.invalid_request", err)
		return
	}
	updated, err := s.store.UpdateProject(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err, "projects.update", "project")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.delete")
	if !ok {
		return
	}
	if err := s.store.DeleteProject(r.Context(), id); err != nil {
		writeStoreError(w, err, "projects.delete", "project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type projectCheckResult struct {
	Passed bool   `json:"passed"`
	Reason string `json:"reason"`
}

func (s *Server) checkProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.check")
	if !ok {
		return
	}
	if _, err := s.store.Project(r.Context(), id); err != nil {
		writeStoreError(w, err, "projects.check", "project")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "projects.check.invalid_request", err)
		return
	}
	result := checkWorkflowTQUsage(string(body))
	writeJSON(w, http.StatusOK, result)
}

func checkWorkflowTQUsage(content string) projectCheckResult {
	normalized := strings.ToLower(content)
	for _, token := range []string{"tq issue ", "tq comment ", "tq project ", "`tq ", "\ntq "} {
		if strings.Contains(normalized, token) {
			return projectCheckResult{Passed: true, Reason: "WORKFLOW.md documents tq command usage"}
		}
	}
	return projectCheckResult{Passed: false, Reason: "WORKFLOW.md does not document tq command usage"}
}

func (s *Server) issues(w http.ResponseWriter, r *http.Request) {
	states, ok := parseIssueStates(w, r)
	if !ok {
		return
	}
	projectID, ok := parseIssueProjectID(w, r)
	if !ok {
		return
	}
	items, err := s.store.IssuesByFilter(r.Context(), store.IssueFilter{States: states, ProjectID: projectID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issues.list.internal_error", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func parseIssueStates(w http.ResponseWriter, r *http.Request) ([]entity.Status, bool) {
	value := r.URL.Query().Get("states")
	if value == "" {
		return nil, true
	}

	parts := strings.Split(value, ",")
	states := make([]entity.Status, 0, len(parts))
	for _, part := range parts {
		state := entity.Status(strings.TrimSpace(part))
		if state == "" {
			continue
		}
		if !entity.IsValidStatus(state) {
			writeError(w, http.StatusBadRequest, "issues.list.invalid_states", errors.New("states contains invalid status"))
			return nil, false
		}
		states = append(states, state)
	}
	return states, true
}

func parseIssueProjectID(w http.ResponseWriter, r *http.Request) (*int64, bool) {
	value := r.URL.Query().Get("project_id")
	if value == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "issues.list.invalid_project_id", errors.New("project_id is invalid"))
		return nil, false
	}
	return &id, true
}

func (s *Server) createIssue(w http.ResponseWriter, r *http.Request) {
	var input entity.CreateIssueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "issues.create.invalid_request", err)
		return
	}
	created, err := s.store.CreateIssue(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "issues.create.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) issue(w http.ResponseWriter, r *http.Request) {
	id, ok := issueID(w, r, "issues.get")
	if !ok {
		return
	}
	item, err := s.store.Issue(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "issues.get", "issue")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) updateIssue(w http.ResponseWriter, r *http.Request) {
	id, ok := issueID(w, r, "issues.update")
	if !ok {
		return
	}
	var input entity.UpdateIssueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "issues.update.invalid_request", err)
		return
	}
	updated, err := s.store.UpdateIssue(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err, "issues.update", "issue")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

type issueStatesInput struct {
	IDs []int64 `json:"ids"`
}

func (s *Server) issueStates(w http.ResponseWriter, r *http.Request) {
	var input issueStatesInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "issues.states.invalid_request", err)
		return
	}
	items, err := s.store.IssueStatesByIDs(r.Context(), input.IDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issues.states.internal_error", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "comments.create")
	if !ok {
		return
	}
	var input entity.CreateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "comments.create.invalid_request", err)
		return
	}
	input.IssueID = issueID
	created, err := s.store.CreateComment(r.Context(), input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "comments.create.issue_not_found", errors.New("issue not found"))
			return
		}
		writeError(w, http.StatusBadRequest, "comments.create.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) comments(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "comments.list")
	if !ok {
		return
	}
	cursor, limit, ok := parseCommentListQuery(w, r)
	if !ok {
		return
	}
	items, err := s.store.CommentsByIssueID(r.Context(), issueID, cursor, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "comments.list.issue_not_found", errors.New("issue not found"))
			return
		}
		writeError(w, http.StatusBadRequest, "comments.list.invalid_input", err)
		return
	}
	var nextCursor *int64
	if len(items) == limit {
		next := items[len(items)-1].ID
		nextCursor = &next
	}
	writeJSONWithMeta(w, http.StatusOK, items, responseMeta{
		"cursor":     cursor,
		"limit":      limit,
		"nextCursor": nextCursor,
	})
}

func (s *Server) updateComment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "comment", "comments.update")
	if !ok {
		return
	}
	var input entity.UpdateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "comments.update.invalid_request", err)
		return
	}
	updated, err := s.store.UpdateComment(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err, "comments.update", "comment")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

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

func parseCommentListQuery(w http.ResponseWriter, r *http.Request) (int64, int, bool) {
	cursor, err := parseOptionalInt64Query(r, "cursor", 0)
	if err != nil || cursor < 0 {
		writeError(w, http.StatusBadRequest, "comments.list.invalid_input", errors.New("cursor is invalid"))
		return 0, 0, false
	}
	limit64, err := parseOptionalInt64Query(r, "limit", 50)
	if err != nil || limit64 < 0 {
		writeError(w, http.StatusBadRequest, "comments.list.invalid_input", errors.New("limit is invalid"))
		return 0, 0, false
	}
	limit := int(limit64)
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return cursor, limit, true
}

func parseOptionalInt64Query(r *http.Request, name string, fallback int64) (int64, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func issueID(w http.ResponseWriter, r *http.Request, action string) (int64, bool) {
	return pathID(w, r, "issue", action)
}

func pathID(w http.ResponseWriter, r *http.Request, resource string, action string) (int64, bool) {
	return pathParamID(w, r, "id", resource, action)
}

func issueIDPath(w http.ResponseWriter, r *http.Request, action string) (int64, bool) {
	return pathParamID(w, r, "issueId", "issue", action)
}

func pathParamID(w http.ResponseWriter, r *http.Request, name string, resource string, action string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, action+".invalid_id", errors.New(resource+" id is invalid"))
		return 0, false
	}
	return id, true
}

func writeStoreError(w http.ResponseWriter, err error, action string, resource string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, action+".not_found", errors.New(resource+" not found"))
		return
	}
	writeError(w, http.StatusBadRequest, action+".invalid_input", err)
}

type responseMeta map[string]any

type successResponse struct {
	Data any          `json:"data"`
	Meta responseMeta `json:"meta"`
}

type errorResponse struct {
	Error responseError `json:"error"`
	Meta  responseMeta  `json:"meta"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	writeRawJSON(w, status, successResponse{Data: value, Meta: responseMeta{}})
}

func writeJSONWithMeta(w http.ResponseWriter, status int, value any, meta responseMeta) {
	writeRawJSON(w, status, successResponse{Data: value, Meta: meta})
}

func writeError(w http.ResponseWriter, status int, code string, err error) {
	writeRawJSON(w, status, errorResponse{
		Error: responseError{Code: code, Message: err.Error()},
		Meta:  responseMeta{},
	})
}

func writeRawJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
