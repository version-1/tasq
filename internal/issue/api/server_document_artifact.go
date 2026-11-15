package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

func (s *Server) createDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	var projectID int64
	var legacyIssueID int64
	if r.PathValue("id") != "" {
		var ok bool
		projectID, ok = pathID(w, r, "project", "document_artifacts.create")
		if !ok {
			return
		}
	} else {
		issueID, ok := issueIDPath(w, r, "document_artifacts.create")
		if !ok {
			return
		}
		issue, err := s.store.Issue(r.Context(), issueID)
		if err != nil {
			writeDocumentArtifactError(w, err, "document_artifacts.create")
			return
		}
		projectID = issue.ProjectID
		legacyIssueID = issueID
	}
	var input entity.CreateDocumentArtifactInput
	if !decodeOneJSON(w, r, "document_artifacts.create", &input) {
		return
	}
	item, err := s.store.CreateDocumentArtifact(r.Context(), projectID, input)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.create")
		return
	}
	if legacyIssueID != 0 {
		if err := s.store.LinkDocumentArtifactToIssue(r.Context(), projectID, item.ID, legacyIssueID); err != nil {
			writeDocumentArtifactError(w, err, "document_artifacts.create")
			return
		}
	}
	writeJSON(w, http.StatusCreated, item)
}
func (s *Server) documentArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") != "" {
		projectID, ok := pathID(w, r, "project", "document_artifacts.list")
		if !ok {
			return
		}
		items, err := s.store.DocumentArtifacts(r.Context(), projectID)
		if err != nil {
			writeDocumentArtifactError(w, err, "document_artifacts.list")
			return
		}
		writeJSON(w, http.StatusOK, items)
		return
	}
	issueID, ok := issueIDPath(w, r, "document_artifacts.list")
	if !ok {
		return
	}
	items, err := s.store.DocumentArtifactsForIssue(r.Context(), issueID)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.list")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
func (s *Server) documentArtifact(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathID(w, r, "project", "document_artifacts.get")
	if !ok {
		return
	}
	item, err := s.store.DocumentArtifact(r.Context(), projectID, r.PathValue("documentId"))
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.get")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) updateDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathID(w, r, "project", "document_artifacts.update")
	if !ok {
		return
	}
	var input entity.UpdateDocumentArtifactInput
	if !decodeOneJSON(w, r, "document_artifacts.update", &input) {
		return
	}
	item, err := s.store.UpdateDocumentArtifact(r.Context(), projectID, r.PathValue("documentId"), input)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.update")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) deleteDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathID(w, r, "project", "document_artifacts.delete")
	if !ok {
		return
	}
	if err := s.store.DeleteDocumentArtifact(r.Context(), projectID, r.PathValue("documentId")); err != nil {
		if errors.Is(err, store.ErrDocumentArtifactStillLinked) {
			writeError(w, http.StatusConflict, "document_artifacts.delete.conflict", err)
			return
		}
		writeDocumentArtifactError(w, err, "document_artifacts.delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) linkDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathID(w, r, "project", "document_artifacts.link")
	if !ok {
		return
	}
	issueID, ok := issueIDPath(w, r, "document_artifacts.link")
	if !ok {
		return
	}
	if err := s.store.LinkDocumentArtifactToIssue(r.Context(), projectID, r.PathValue("documentId"), issueID); err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.link")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) unlinkDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	projectID, ok := pathID(w, r, "project", "document_artifacts.unlink")
	if !ok {
		return
	}
	issueID, ok := issueIDPath(w, r, "document_artifacts.unlink")
	if !ok {
		return
	}
	if err := s.store.UnlinkDocumentArtifactFromIssue(r.Context(), projectID, r.PathValue("documentId"), issueID); err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.unlink")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) issueDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "document_artifacts.get")
	if !ok {
		return
	}
	items, err := s.store.DocumentArtifactsForIssue(r.Context(), issueID)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.get")
		return
	}
	for _, item := range items {
		if item.ID == r.PathValue("documentId") {
			full, err := s.store.DocumentArtifact(r.Context(), item.ProjectID, item.ID)
			if err != nil {
				writeDocumentArtifactError(w, err, "document_artifacts.get")
				return
			}
			writeJSON(w, http.StatusOK, full)
			return
		}
	}
	writeError(w, http.StatusNotFound, "document_artifacts.get.not_found", errors.New("document artifact not linked to issue"))
}
func (s *Server) updateIssueDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "document_artifacts.update")
	if !ok {
		return
	}
	issue, err := s.store.Issue(r.Context(), issueID)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.update")
		return
	}
	items, err := s.store.DocumentArtifactsForIssue(r.Context(), issueID)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.update")
		return
	}
	for _, item := range items {
		if item.ID == r.PathValue("documentId") {
			if item.ProjectID != issue.ProjectID {
				writeError(w, http.StatusForbidden, "document_artifacts.update.owner_project_required", errors.New("document artifact is owned by another project"))
				return
			}
			var input entity.UpdateDocumentArtifactInput
			if !decodeOneJSON(w, r, "document_artifacts.update", &input) {
				return
			}
			updated, err := s.store.UpdateDocumentArtifact(r.Context(), item.ProjectID, item.ID, input)
			if err != nil {
				writeDocumentArtifactError(w, err, "document_artifacts.update")
				return
			}
			writeJSON(w, http.StatusOK, updated)
			return
		}
	}
	writeError(w, http.StatusNotFound, "document_artifacts.update.not_found", errors.New("document artifact not linked to issue"))
}
func decodeOneJSON(w http.ResponseWriter, r *http.Request, action string, value any) bool {
	// The decoded Markdown body is limited by the domain layer. JSON escaping
	// can expand a valid 100 kB body by up to six times, so leave room for the
	// encoded envelope while still bounding the request before reading it.
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, action+".invalid_request", err)
		return false
	}
	if !utf8.Valid(raw) {
		writeError(w, http.StatusBadRequest, action+".invalid_request", errors.New("request body is invalid UTF-8"))
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, action+".invalid_request", err)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, action+".invalid_request", errors.New("request body must contain exactly one JSON object"))
		return false
	}
	return true
}
func writeDocumentArtifactError(w http.ResponseWriter, err error, action string) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, action+".not_found", errors.New("issue or document artifact not found"))
		return
	}
	writeError(w, http.StatusBadRequest, action+".invalid_input", err)
}
