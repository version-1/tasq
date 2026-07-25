package api

import (
	"encoding/json"
	"net/http"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

func (s *Server) issues(w http.ResponseWriter, r *http.Request) {
	states, ok := parseIssueStates(w, r)
	if !ok {
		return
	}
	projectID, ok := parseIssueProjectID(w, r)
	if !ok {
		return
	}
	projectIDs, ok := parseIssueProjectIDs(w, r)
	if !ok {
		return
	}
	priorities, ok := parseIssuePriorities(w, r)
	if !ok {
		return
	}
	assignee := parseIssueAssignee(r)
	search := parseIssueSearch(r)
	limit, offset, ok := parseIssuePagination(w, r)
	if !ok {
		return
	}
	sortBy, sortDirection, ok := parseIssueSort(w, r)
	if !ok {
		return
	}
	list, err := s.store.IssuesPageByFilter(r.Context(), store.IssueFilter{
		States:        states,
		ProjectID:     projectID,
		ProjectIDs:    projectIDs,
		Priorities:    priorities,
		Assignee:      assignee,
		Search:        search,
		Limit:         limit,
		Offset:        offset,
		SortBy:        sortBy,
		SortDirection: sortDirection,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issues.list.internal_error", err)
		return
	}
	writeJSONWithMeta(w, http.StatusOK, list.Issues, responseMeta{
		"limit":      limit,
		"offset":     offset,
		"total":      list.Total,
		"nextOffset": nextIssueOffset(offset, limit, list.Total),
	})
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

func (s *Server) queue(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectIDQuery(w, r, "queue.list")
	if !ok {
		return
	}
	queue, err := s.store.Queue(r.Context(), store.IssueFilter{ProjectID: projectID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "queue.list.internal_error", err)
		return
	}
	writeJSON(w, http.StatusOK, queue)
}
