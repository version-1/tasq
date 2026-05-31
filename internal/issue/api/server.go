package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

type Server struct {
	store *store.Store
}

func NewServer(store *store.Store) *Server {
	return &Server{store: store}
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
	mux.HandleFunc("GET /api/v1/workspaces", s.workspaces)
	mux.HandleFunc("POST /api/v1/workspaces", s.createWorkspace)
	mux.HandleFunc("GET /api/v1/workspaces/{id}", s.workspace)
	mux.HandleFunc("PATCH /api/v1/workspaces/{id}", s.updateWorkspace)
	mux.HandleFunc("DELETE /api/v1/workspaces/{id}", s.deleteWorkspace)
	mux.HandleFunc("GET /api/v1/issues", s.issues)
	mux.HandleFunc("POST /api/v1/issues", s.createIssue)
	mux.HandleFunc("POST /api/v1/issues/states", s.issueStates)
	mux.HandleFunc("GET /api/v1/issues/{id}", s.issue)
	mux.HandleFunc("PATCH /api/v1/issues/{id}", s.updateIssue)
	mux.HandleFunc("POST /api/v1/work-items/claim", s.claimWorkItem)
	mux.HandleFunc("POST /api/v1/work-items/renew-lease", s.renewWorkItemLease)
	mux.HandleFunc("POST /api/v1/orchestrator-events", s.receiveRunEvent)
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

func (s *Server) workspaces(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Workspaces(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "workspaces.list.internal_error", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var input entity.CreateWorkspaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "workspaces.create.invalid_request", err)
		return
	}
	created, err := s.store.CreateWorkspace(r.Context(), input)
	if err != nil {
		writeStoreError(w, err, "workspaces.create", "workspace")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) workspace(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "workspace", "workspaces.get")
	if !ok {
		return
	}
	item, err := s.store.Workspace(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "workspaces.get", "workspace")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "workspace", "workspaces.update")
	if !ok {
		return
	}
	var input entity.UpdateWorkspaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "workspaces.update.invalid_request", err)
		return
	}
	updated, err := s.store.UpdateWorkspace(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err, "workspaces.update", "workspace")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "workspace", "workspaces.delete")
	if !ok {
		return
	}
	if err := s.store.DeleteWorkspace(r.Context(), id); err != nil {
		writeStoreError(w, err, "workspaces.delete", "workspace")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) issues(w http.ResponseWriter, r *http.Request) {
	states, ok := parseIssueStates(w, r)
	if !ok {
		return
	}
	items, err := s.store.IssuesByStates(r.Context(), states)
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

func (s *Server) claimWorkItem(w http.ResponseWriter, r *http.Request) {
	var input entity.ClaimWorkItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "work_items.claim.invalid_request", err)
		return
	}
	item, err := s.store.ClaimWorkItem(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "work_items.claim.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusOK, entity.ClaimWorkItemOutput{WorkItem: item})
}

func (s *Server) renewWorkItemLease(w http.ResponseWriter, r *http.Request) {
	var input entity.RenewWorkItemLeaseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "work_items.renew_lease.invalid_request", err)
		return
	}
	item, err := s.store.RenewWorkItemLease(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "work_items.renew_lease.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusOK, entity.RenewWorkItemLeaseOutput{WorkItem: item})
}

func (s *Server) receiveRunEvent(w http.ResponseWriter, r *http.Request) {
	var input entity.RunEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "orchestrator_events.receive.invalid_request", err)
		return
	}
	if err := s.store.ReceiveRunEvent(r.Context(), input); err != nil {
		writeError(w, http.StatusBadRequest, "orchestrator_events.receive.invalid_input", err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func issueID(w http.ResponseWriter, r *http.Request, action string) (int64, bool) {
	return pathID(w, r, "issue", action)
}

func pathID(w http.ResponseWriter, r *http.Request, resource string, action string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
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

type responseMeta struct{}

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
