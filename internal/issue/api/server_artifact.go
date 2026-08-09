package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (s *Server) upsertArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "artifacts.upsert")
	if !ok {
		return
	}
	var input entity.UpsertArtifactInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "artifacts.upsert.invalid_request", err)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "artifacts.upsert.invalid_request", errors.New("request body must contain exactly one JSON object"))
		return
	}
	item, err := s.store.UpsertArtifact(r.Context(), issueID, entity.ArtifactType(r.PathValue("type")), input)
	if err != nil {
		writeStoreError(w, err, "artifacts.upsert", "artifact")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) deleteArtifact(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "artifacts.delete")
	if !ok {
		return
	}
	if err := s.store.DeleteArtifact(r.Context(), issueID, entity.ArtifactType(r.PathValue("type"))); err != nil {
		writeStoreError(w, err, "artifacts.delete", "artifact")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
