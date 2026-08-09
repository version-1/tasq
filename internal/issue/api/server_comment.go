package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

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
	cursor, limit, direction, ok := parseCommentListQuery(w, r)
	if !ok {
		return
	}
	items, err := s.store.CommentsPageByIssueID(r.Context(), issueID, cursor, limit, direction)
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
		"direction":  direction,
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

func (s *Server) createChangeRequest(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "change_requests.create")
	if !ok {
		return
	}
	var input entity.CreateChangeRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "change_requests.create.invalid_request", err)
		return
	}
	input.IssueID = issueID
	created, err := s.store.CreateChangeRequest(r.Context(), input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "change_requests.create.issue_not_found", errors.New("issue not found"))
			return
		}
		writeError(w, http.StatusBadRequest, "change_requests.create.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) changeRequests(w http.ResponseWriter, r *http.Request) {
	issueID, ok := issueIDPath(w, r, "change_requests.list")
	if !ok {
		return
	}
	status, limit, ok := parseChangeRequestListQuery(w, r)
	if !ok {
		return
	}
	items, err := s.store.ChangeRequestsByIssueID(r.Context(), issueID, status, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "change_requests.list.issue_not_found", errors.New("issue not found"))
			return
		}
		writeError(w, http.StatusBadRequest, "change_requests.list.invalid_input", err)
		return
	}
	writeJSONWithMeta(w, http.StatusOK, items, responseMeta{
		"limit":  limit,
		"status": status,
	})
}

func (s *Server) changeRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "change request", "change_requests.get")
	if !ok {
		return
	}
	item, err := s.store.ChangeRequest(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "change_requests.get", "change request")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) updateChangeRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "change request", "change_requests.update")
	if !ok {
		return
	}
	var input entity.UpdateChangeRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "change_requests.update.invalid_request", err)
		return
	}
	updated, err := s.store.UpdateChangeRequest(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err, "change_requests.update", "change request")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) cancelChangeRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "change request", "change_requests.cancel")
	if !ok {
		return
	}
	updated, err := s.store.CancelChangeRequest(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "change_requests.cancel", "change request")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
