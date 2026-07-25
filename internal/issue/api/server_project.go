package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"gopkg.in/yaml.v3"
)

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
	if s.projectRunDataDeleter != nil {
		issueIDs, err := s.store.ProjectIssueIDs(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "projects.delete.project_issue_ids_failed", err)
			return
		}
		if err := s.projectRunDataDeleter.DeleteProjectIssueData(r.Context(), issueIDs); err != nil {
			if errors.Is(err, runstore.ErrProjectHasRunningRuns) {
				writeError(w, http.StatusConflict, "projects.delete.running_runs", err)
				return
			}
			writeError(w, http.StatusInternalServerError, "projects.delete.orchestrator_cleanup_failed", err)
			return
		}
	}
	if err := s.store.DeleteProject(r.Context(), id); err != nil {
		writeStoreError(w, err, "projects.delete", "project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type upsertProjectWorkflowRequest struct {
	Frontmatter json.RawMessage `json:"frontmatter"`
	Body        *string         `json:"body"`
	Checksum    string          `json:"checksum"`
}

func (s *Server) upsertProjectWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.workflow.upsert")
	if !ok {
		return
	}
	var request upsertProjectWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "projects.workflow.upsert.invalid_request", err)
		return
	}
	frontmatterJSON, err := compactJSONObject(request.Frontmatter)
	if err != nil {
		writeError(w, http.StatusBadRequest, "projects.workflow.upsert.invalid_input", err)
		return
	}
	if request.Body == nil {
		writeError(w, http.StatusBadRequest, "projects.workflow.upsert.invalid_input", errors.New("body is required"))
		return
	}
	workflow, err := s.store.UpsertProjectWorkflow(r.Context(), entity.UpsertProjectWorkflowInput{
		ProjectID:       id,
		FrontmatterJSON: frontmatterJSON,
		Body:            *request.Body,
		Checksum:        request.Checksum,
	})
	if err != nil {
		writeStoreError(w, err, "projects.workflow.upsert", "project workflow")
		return
	}
	writeJSON(w, http.StatusOK, workflow)
}

func (s *Server) projectWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.workflow")
	if !ok {
		return
	}
	item, err := s.store.ProjectWorkflow(r.Context(), id)
	if err != nil {
		writeStoreError(w, err, "projects.workflow", "workflow")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) deleteProjectWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "project", "projects.workflow.delete")
	if !ok {
		return
	}
	if err := s.store.DeleteProjectWorkflow(r.Context(), id); err != nil {
		writeStoreError(w, err, "projects.workflow.delete", "project workflow")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func compactJSONObject(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", errors.New("frontmatter is required")
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return "", errors.New("frontmatter must be a JSON object")
	}
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		return "", err
	}
	return buffer.String(), nil
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
	if !workflowDisablesTaskWorkPrompt(content) {
		return projectCheckResult{Passed: true, Reason: "WORKFLOW.md uses the default tq task work prompt"}
	}
	return projectCheckResult{Passed: false, Reason: "WORKFLOW.md does not document tq command usage"}
}

func workflowDisablesTaskWorkPrompt(content string) bool {
	if !strings.HasPrefix(content, "---\n") {
		return false
	}
	frontMatter, _, ok := strings.Cut(strings.TrimPrefix(content, "---\n"), "\n---")
	if !ok {
		return false
	}
	var parsed struct {
		Tasq struct {
			TaskWorkPrompt *bool `yaml:"task_work_prompt"`
		} `yaml:"tasq"`
	}
	if err := yaml.Unmarshal([]byte(frontMatter), &parsed); err != nil {
		return false
	}
	return parsed.Tasq.TaskWorkPrompt != nil && !*parsed.Tasq.TaskWorkPrompt
}
