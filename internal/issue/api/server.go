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
	projectDeletionGuard  projectDeletionGuard
	projectRunDataDeleter projectRunDataDeleter
}

type projectRunDataDeleter interface {
	DeleteProjectIssueData(ctx context.Context, issueIDs []int64) error
}

type projectDeletionGuard interface {
	CheckProjectDeletion(ctx context.Context) error
}

type projectDeletionGuardFunc func(ctx context.Context) error

func (f projectDeletionGuardFunc) CheckProjectDeletion(ctx context.Context) error { return f(ctx) }

// SetProjectDeletionGuard blocks project deletion when external run history
// cannot be safely inspected without opening the legacy orchestrator database.
func (s *Server) SetProjectDeletionGuard(check func(context.Context) error) {
	if check == nil {
		s.projectDeletionGuard = nil
		return
	}
	s.projectDeletionGuard = projectDeletionGuardFunc(check)
}

func NewServer(issueStore *store.Store) *Server {
	return &Server{store: issueStore}
}

func NewServerWithAttachmentStorage(issueStore *store.Store, attachmentStorage *store.AttachmentStorage) *Server {
	issueStore.SetAttachmentStorage(attachmentStorage)
	return &Server{store: issueStore, attachmentStorage: attachmentStorage}
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
	mux.HandleFunc("POST /api/v1/issues/{id}/auto-ok-result", s.completeAutoOK)
	mux.HandleFunc("PUT /api/v1/issues/{issueId}/artifacts/{type}", s.upsertArtifact)
	mux.HandleFunc("DELETE /api/v1/issues/{issueId}/artifacts/{type}", s.deleteArtifact)
	mux.HandleFunc("GET /api/v1/issues/{issueId}/artifacts/documents", s.documentArtifacts)
	mux.HandleFunc("POST /api/v1/issues/{issueId}/artifacts/documents", s.createDocumentArtifact)
	mux.HandleFunc("GET /api/v1/issues/{issueId}/artifacts/documents/{documentId}", s.documentArtifact)
	mux.HandleFunc("PATCH /api/v1/issues/{issueId}/artifacts/documents/{documentId}", s.updateDocumentArtifact)
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
