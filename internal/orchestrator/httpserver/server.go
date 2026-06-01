package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

type Store interface {
	ActiveRuns(ctx context.Context) ([]run.Run, error)
	RunByIssueID(ctx context.Context, issueID int64) (run.Run, error)
	RunnerEvents(ctx context.Context, runID string, limit int) ([]run.RunnerEvent, error)
	TokensByRunIDs(ctx context.Context, runIDs []string) (map[string]run.TokenSummary, error)
}

type Refresher interface {
	RequestRefresh()
}

type Server struct {
	store     Store
	refresher Refresher
}

func New(store Store, refresher Refresher) *Server {
	return &Server{store: store, refresher: refresher}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/state", s.state)
	mux.HandleFunc("POST /api/v1/refresh", s.refresh)
	mux.HandleFunc("GET /api/v1/{issue_identifier}", s.issue)
	return mux
}

func ListenAndServe(ctx context.Context, port int, handler http.Handler) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return "", err
	}
	server := &http.Server{Handler: handler}
	errCh := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	select {
	case err := <-errCh:
		return "", err
	default:
	}
	return "http://" + listener.Addr().String(), nil
}

type stateResponse struct {
	GeneratedAt time.Time   `json:"generated_at"`
	Counts      stateCounts `json:"counts"`
	Running     []runRow    `json:"running"`
	Retrying    []retryRow  `json:"retrying"`
	CodexTotals any         `json:"codex_totals"`
	RateLimits  any         `json:"rate_limits"`
}

type stateCounts struct {
	Running  int `json:"running"`
	Retrying int `json:"retrying"`
}

type runRow struct {
	IssueID         string       `json:"issue_id"`
	IssueIdentifier string       `json:"issue_identifier"`
	State           string       `json:"state"`
	SessionID       string       `json:"session_id"`
	TurnCount       int          `json:"turn_count"`
	LastEvent       string       `json:"last_event"`
	LastMessage     string       `json:"last_message"`
	StartedAt       time.Time    `json:"started_at"`
	LastEventAt     *time.Time   `json:"last_event_at"`
	Tokens          tokenSummary `json:"tokens"`
}

type tokenSummary struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

type retryRow struct{}

type issueResponse struct {
	IssueIdentifier string           `json:"issue_identifier"`
	IssueID         string           `json:"issue_id"`
	Status          string           `json:"status"`
	Workspace       workspacePayload `json:"workspace"`
	Attempts        attemptsPayload  `json:"attempts"`
	Running         *runRow          `json:"running"`
	Retry           any              `json:"retry"`
	Logs            logsPayload      `json:"logs"`
	RecentEvents    []eventRow       `json:"recent_events"`
	LastError       any              `json:"last_error"`
	Tracked         map[string]any   `json:"tracked"`
}

type workspacePayload struct {
	Path string `json:"path"`
}

type attemptsPayload struct {
	RestartCount        int `json:"restart_count"`
	CurrentRetryAttempt int `json:"current_retry_attempt"`
}

type logsPayload struct {
	CodexSessionLogs []any `json:"codex_session_logs"`
}

type eventRow struct {
	At      time.Time `json:"at"`
	Event   string    `json:"event"`
	Message string    `json:"message"`
}

func (s *Server) state(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ActiveRuns(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "state_query_failed", err.Error())
		return
	}
	runIDs := make([]string, 0, len(runs))
	for _, storedRun := range runs {
		runIDs = append(runIDs, storedRun.RunID)
	}
	tokensByRunID, err := s.store.TokensByRunIDs(r.Context(), runIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "state_query_failed", err.Error())
		return
	}
	rows := make([]runRow, 0, len(runs))
	totals := tokenSummary{}
	for _, storedRun := range runs {
		row, err := s.runRow(r.Context(), storedRun)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "state_query_failed", err.Error())
			return
		}
		row.Tokens = tokenSummaryFromRun(tokensByRunID[storedRun.RunID])
		totals.InputTokens += row.Tokens.InputTokens
		totals.OutputTokens += row.Tokens.OutputTokens
		totals.TotalTokens += row.Tokens.TotalTokens
		rows = append(rows, row)
	}
	response := stateResponse{
		GeneratedAt: time.Now().UTC(),
		Counts: stateCounts{
			Running:  len(rows),
			Retrying: 0,
		},
		Running:     rows,
		Retrying:    []retryRow{},
		CodexTotals: totals,
		RateLimits:  nil,
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) issue(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("issue_identifier")
	issueID, ok := parseIssueIdentifier(identifier)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_issue_identifier", "issue identifier must use issue-<id>")
		return
	}
	storedRun, err := s.store.RunByIssueID(r.Context(), issueID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "issue_not_found", "issue not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue_query_failed", err.Error())
		return
	}
	events, err := s.store.RunnerEvents(r.Context(), storedRun.RunID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "issue_query_failed", err.Error())
		return
	}
	eventRows := make([]eventRow, 0, len(events))
	for _, event := range events {
		eventRows = append(eventRows, eventRow{
			At:      event.OccurredAt,
			Event:   event.EventType,
			Message: event.Message,
		})
	}
	var running *runRow
	if storedRun.Status == run.StatusQueued || storedRun.Status == run.StatusRunning {
		row, err := s.runRowFromEvents(storedRun, events)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "issue_query_failed", err.Error())
			return
		}
		running = &row
	}
	var lastError any
	if storedRun.Error != "" {
		lastError = storedRun.Error
	}
	response := issueResponse{
		IssueIdentifier: identifier,
		IssueID:         strconv.FormatInt(storedRun.IssueID, 10),
		Status:          string(storedRun.Status),
		Workspace:       workspacePayload{Path: storedRun.Workspace},
		Attempts: attemptsPayload{
			RestartCount:        max(storedRun.Attempt-1, 0),
			CurrentRetryAttempt: storedRun.Attempt,
		},
		Running:      running,
		Retry:        nil,
		Logs:         logsPayload{CodexSessionLogs: []any{}},
		RecentEvents: eventRows,
		LastError:    lastError,
		Tracked:      map[string]any{},
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	if s.refresher == nil {
		writeError(w, http.StatusServiceUnavailable, "refresh_unavailable", "worker refresh is unavailable")
		return
	}
	s.refresher.RequestRefresh()
	writeJSON(w, http.StatusAccepted, map[string]any{
		"queued":       true,
		"coalesced":    false,
		"requested_at": time.Now().UTC(),
		"operations":   []string{"poll", "reconcile"},
	})
}

func (s *Server) runRow(ctx context.Context, storedRun run.Run) (runRow, error) {
	events, err := s.store.RunnerEvents(ctx, storedRun.RunID, 100)
	if err != nil {
		return runRow{}, err
	}
	return s.runRowFromEvents(storedRun, events)
}

func (s *Server) runRowFromEvents(storedRun run.Run, events []run.RunnerEvent) (runRow, error) {
	lastEvent := ""
	lastMessage := ""
	var lastEventAt *time.Time
	turnCount := 0
	sessionID := ""
	for _, event := range events {
		lastEvent = event.EventType
		lastMessage = event.Message
		eventAt := event.OccurredAt
		lastEventAt = &eventAt
		if event.EventType == "turn_started" {
			turnCount++
		}
		if sessionID == "" && event.EventType == "session_started" {
			sessionID = strings.TrimPrefix(event.Message, "thread_id=")
		}
	}
	return runRow{
		IssueID:         strconv.FormatInt(storedRun.IssueID, 10),
		IssueIdentifier: issueIdentifier(storedRun.IssueID),
		State:           string(storedRun.Status),
		SessionID:       sessionID,
		TurnCount:       turnCount,
		LastEvent:       lastEvent,
		LastMessage:     lastMessage,
		StartedAt:       storedRun.CreatedAt,
		LastEventAt:     lastEventAt,
		Tokens:          tokenSummary{},
	}, nil
}

func parseIssueIdentifier(identifier string) (int64, bool) {
	raw, ok := strings.CutPrefix(identifier, "issue-")
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func issueIdentifier(issueID int64) string {
	return fmt.Sprintf("issue-%d", issueID)
}

func tokenSummaryFromRun(summary run.TokenSummary) tokenSummary {
	return tokenSummary{
		InputTokens:  summary.InputTokens,
		OutputTokens: summary.OutputTokens,
		TotalTokens:  summary.TotalTokens,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
