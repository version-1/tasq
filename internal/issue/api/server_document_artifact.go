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
)

func (s *Server) createDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "document_artifacts.create")
	if !ok {
		return
	}
	var input entity.CreateDocumentArtifactInput
	if !decodeOneJSON(w, r, "document_artifacts.create", &input) {
		return
	}
	item, err := s.store.CreateDocumentArtifact(r.Context(), issueID, input)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.create")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
func (s *Server) documentArtifacts(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "document_artifacts.list")
	if !ok {
		return
	}
	items, err := s.store.DocumentArtifacts(r.Context(), issueID)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.list")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
func (s *Server) documentArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "document_artifacts.get")
	if !ok {
		return
	}
	item, err := s.store.DocumentArtifact(r.Context(), issueID, r.PathValue("documentId"))
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.get")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func (s *Server) updateDocumentArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "document_artifacts.update")
	if !ok {
		return
	}
	var input entity.UpdateDocumentArtifactInput
	if !decodeOneJSON(w, r, "document_artifacts.update", &input) {
		return
	}
	item, err := s.store.UpdateDocumentArtifact(r.Context(), issueID, r.PathValue("documentId"), input)
	if err != nil {
		writeDocumentArtifactError(w, err, "document_artifacts.update")
		return
	}
	writeJSON(w, http.StatusOK, item)
}
func decodeOneJSON(w http.ResponseWriter, r *http.Request, action string, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 400000)
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
