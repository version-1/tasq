package api

import (
	"context"
	"net/http"

	"github.com/version-1/tasq/internal/issue/store"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
)

type Server struct {
	store                 *store.Store
	attachmentStorage     *store.AttachmentStorage
	projectRunDataDeleter projectRunDataDeleter
}

type projectRunDataDeleter interface {
	DeleteProjectIssueData(ctx context.Context, issueIDs []int64) error
}

func NewServer(issueStore *store.Store) *Server {
	return &Server{store: issueStore}
}

func NewServerWithAttachmentStorage(issueStore *store.Store, attachmentStorage *store.AttachmentStorage) *Server {
	issueStore.SetAttachmentStorage(attachmentStorage)
	return &Server{store: issueStore, attachmentStorage: attachmentStorage}
}

func NewServerWithOrchestratorStore(issueStore *store.Store, orchestratorStore *runstore.Store) *Server {
	return &Server{store: issueStore, projectRunDataDeleter: orchestratorStore}
}

func NewServerWithStores(issueStore *store.Store, attachmentStorage *store.AttachmentStorage, orchestratorStore *runstore.Store) *Server {
	issueStore.SetAttachmentStorage(attachmentStorage)
	return &Server{store: issueStore, attachmentStorage: attachmentStorage, projectRunDataDeleter: orchestratorStore}
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
	mux.HandleFunc("GET /api/v1/projects/{id}/workflow", s.projectWorkflow)
	mux.HandleFunc("PUT /api/v1/projects/{id}/workflow", s.upsertProjectWorkflow)
	mux.HandleFunc("DELETE /api/v1/projects/{id}/workflow", s.deleteProjectWorkflow)
	mux.HandleFunc("POST /api/v1/projects/{id}/check", s.checkProject)
	mux.HandleFunc("GET /api/v1/issues", s.issues)
	mux.HandleFunc("POST /api/v1/issues", s.createIssue)
	mux.HandleFunc("POST /api/v1/issues/states", s.issueStates)
	mux.HandleFunc("GET /api/v1/queue", s.queue)
	mux.HandleFunc("GET /api/v1/issues/{id}", s.issue)
	mux.HandleFunc("PATCH /api/v1/issues/{id}", s.updateIssue)
	mux.HandleFunc("GET /api/v1/issues/{issueId}/comments", s.comments)
	mux.HandleFunc("POST /api/v1/issues/{issueId}/comments", s.createComment)
	mux.HandleFunc("PATCH /api/v1/comments/{id}", s.updateComment)
	mux.HandleFunc("GET /api/v1/issues/{issueId}/change-requests", s.changeRequests)
	mux.HandleFunc("POST /api/v1/issues/{issueId}/change-requests", s.createChangeRequest)
	mux.HandleFunc("GET /api/v1/change-requests/{id}", s.changeRequest)
	mux.HandleFunc("PATCH /api/v1/change-requests/{id}", s.updateChangeRequest)
	mux.HandleFunc("POST /api/v1/change-requests/{id}/cancel", s.cancelChangeRequest)
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
