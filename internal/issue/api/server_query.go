package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

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
	return parseProjectIDQuery(w, r, "issues.list")
}

func parseIssueProjectIDs(w http.ResponseWriter, r *http.Request) ([]int64, bool) {
	value := r.URL.Query().Get("project_ids")
	if value == "" {
		return nil, true
	}
	parts := strings.Split(value, ",")
	projectIDs := make([]int64, 0, len(parts))
	for _, part := range parts {
		raw := strings.TrimSpace(part)
		if raw == "" {
			continue
		}
		projectID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || projectID <= 0 {
			writeError(w, http.StatusBadRequest, "issues.list.invalid_project_ids", errors.New("project_ids is invalid"))
			return nil, false
		}
		projectIDs = append(projectIDs, projectID)
	}
	return projectIDs, true
}

func parseIssuePriorities(w http.ResponseWriter, r *http.Request) ([]entity.Priority, bool) {
	value := r.URL.Query().Get("priorities")
	if value == "" {
		return nil, true
	}
	parts := strings.Split(value, ",")
	priorities := make([]entity.Priority, 0, len(parts))
	for _, part := range parts {
		priority := entity.Priority(strings.TrimSpace(part))
		if priority == "" {
			continue
		}
		if !entity.IsValidPriority(priority) {
			writeError(w, http.StatusBadRequest, "issues.list.invalid_priorities", errors.New("priorities contains invalid priority"))
			return nil, false
		}
		priorities = append(priorities, priority)
	}
	return priorities, true
}

func parseIssueAssignee(r *http.Request) *string {
	value := strings.TrimSpace(r.URL.Query().Get("assignee"))
	if value == "" {
		return nil
	}
	return &value
}

func parseIssueSearch(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get("search"))
}

func parseIssueSort(w http.ResponseWriter, r *http.Request) (store.IssueSortBy, store.SortDirection, bool) {
	sortBy := store.IssueSortBy(strings.TrimSpace(r.URL.Query().Get("sort_by")))
	if sortBy == "" {
		sortBy = store.IssueSortByUpdatedAt
	}
	switch sortBy {
	case store.IssueSortByID, store.IssueSortByPriority, store.IssueSortByCreatedAt, store.IssueSortByUpdatedAt:
	default:
		writeError(w, http.StatusBadRequest, "issues.list.invalid_sort_by", errors.New("sort_by is invalid"))
		return "", "", false
	}

	direction := store.SortDirection(strings.TrimSpace(r.URL.Query().Get("sort_direction")))
	if direction == "" {
		direction = store.SortDirectionDesc
	}
	switch direction {
	case store.SortDirectionAsc, store.SortDirectionDesc:
	default:
		writeError(w, http.StatusBadRequest, "issues.list.invalid_sort_direction", errors.New("sort_direction is invalid"))
		return "", "", false
	}

	return sortBy, direction, true
}

func parseIssuePagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, ok := parsePositiveIntQuery(w, r, "limit", 0, 50, "issues.list.invalid_limit")
	if !ok {
		return 0, 0, false
	}
	offset, ok := parseNonNegativeIntQuery(w, r, "offset", 0, "issues.list.invalid_offset")
	if !ok {
		return 0, 0, false
	}
	if limit == 0 {
		return 0, 0, true
	}
	return limit, offset, true
}

func parsePositiveIntQuery(w http.ResponseWriter, r *http.Request, name string, fallback int, max int, code string) (int, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 || parsed > max {
		writeError(w, http.StatusBadRequest, code, fmt.Errorf("%s is invalid", name))
		return 0, false
	}
	return parsed, true
}

func parseNonNegativeIntQuery(w http.ResponseWriter, r *http.Request, name string, fallback int, code string) (int, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		writeError(w, http.StatusBadRequest, code, fmt.Errorf("%s is invalid", name))
		return 0, false
	}
	return parsed, true
}

func nextIssueOffset(offset int, limit int, total int) *int {
	if limit <= 0 {
		return nil
	}
	next := offset + limit
	if next >= total {
		return nil
	}
	return &next
}

func parseProjectIDQuery(w http.ResponseWriter, r *http.Request, action string) (*int64, bool) {
	value := r.URL.Query().Get("project_id")
	if value == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, action+".invalid_project_id", errors.New("project_id is invalid"))
		return nil, false
	}
	return &id, true
}

func parseCommentListQuery(w http.ResponseWriter, r *http.Request) (int64, int, store.CommentDirection, bool) {
	cursor, err := parseOptionalInt64Query(r, "cursor", 0)
	if err != nil || cursor < 0 {
		writeError(w, http.StatusBadRequest, "comments.list.invalid_input", errors.New("cursor is invalid"))
		return 0, 0, "", false
	}
	limit64, err := parseOptionalInt64Query(r, "limit", 50)
	if err != nil || limit64 < 0 {
		writeError(w, http.StatusBadRequest, "comments.list.invalid_input", errors.New("limit is invalid"))
		return 0, 0, "", false
	}
	limit := int(limit64)
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	direction := store.CommentDirection(strings.TrimSpace(r.URL.Query().Get("direction")))
	if direction == "" {
		direction = store.CommentDirectionForward
	}
	if direction != store.CommentDirectionForward && direction != store.CommentDirectionBackward {
		writeError(w, http.StatusBadRequest, "comments.list.invalid_input", errors.New("direction is invalid"))
		return 0, 0, "", false
	}
	return cursor, limit, direction, true
}

func parseChangeRequestListQuery(w http.ResponseWriter, r *http.Request) (entity.ChangeRequestStatus, int, bool) {
	status := entity.ChangeRequestStatus(strings.TrimSpace(r.URL.Query().Get("status")))
	if status != "" && !entity.IsValidChangeRequestStatus(status) {
		writeError(w, http.StatusBadRequest, "change_requests.list.invalid_input", errors.New("status is invalid"))
		return "", 0, false
	}
	limit64, err := parseOptionalInt64Query(r, "limit", 50)
	if err != nil || limit64 < 0 {
		writeError(w, http.StatusBadRequest, "change_requests.list.invalid_input", errors.New("limit is invalid"))
		return "", 0, false
	}
	limit := int(limit64)
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return status, limit, true
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
