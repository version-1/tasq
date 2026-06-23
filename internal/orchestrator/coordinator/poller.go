package coordinator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

type Tracker interface {
	Issue(ctx context.Context, id int64) (entity.Issue, error)
	Project(ctx context.Context, id int64) (entity.Project, error)
	Queue(ctx context.Context) (entity.Queue, error)
}

type Store interface {
	ActiveRuns(ctx context.Context) ([]run.Run, error)
	RunByIssueID(ctx context.Context, issueID int64) (run.Run, error)
	CreateRun(ctx context.Context, input runstore.CreateRunInput) (run.Run, error)
	RecordRunnerEvent(ctx context.Context, runID string, eventType string, message string, payloadJSON string) error
	RecordWorkspaceSetupFailure(ctx context.Context, issueID int64, workspaceKey string, path string, errText string) error
	UpsertWorkspaceMetadata(ctx context.Context, input runstore.WorkspaceMetadataInput) error
}

type WorkspaceManager interface {
	CreateForIssueInProjectLocationWithConfig(identifier string, projectLocation string, root string, hooks workspace.HookConfig) (workspace.Workspace, error)
}

type Poller struct {
	tracker        Tracker
	store          Store
	workspaces     WorkspaceManager
	workflow       WorkflowResolver
	dispatcher     PollDispatcher
	interval       time.Duration
	maxActiveRuns  int
	orchestratorID string
	refreshCh      chan struct{}
}

type PollDispatcher interface {
	Dispatch(ctx context.Context, activeRuns []run.Run) error
}

type PollerConfig struct {
	Tracker        Tracker
	Store          Store
	Workspaces     WorkspaceManager
	Workflow       WorkflowResolver
	Dispatcher     PollDispatcher
	Interval       time.Duration
	MaxActiveRuns  int
	OrchestratorID string
}

func NewPoller(config PollerConfig) (*Poller, error) {
	if config.Tracker == nil {
		return nil, errors.New("tracker is required")
	}
	if config.Store == nil {
		return nil, errors.New("store is required")
	}
	if config.Workspaces == nil {
		return nil, errors.New("workspace manager is required")
	}
	if config.Workflow == nil {
		return nil, errors.New("workflow resolver is required")
	}
	if config.Interval <= 0 {
		return nil, errors.New("poll interval must be positive")
	}
	if config.MaxActiveRuns <= 0 {
		return nil, errors.New("max active runs must be positive")
	}
	if config.OrchestratorID == "" {
		config.OrchestratorID = "tasq-orchestrator"
	}
	return &Poller{
		tracker:        config.Tracker,
		store:          config.Store,
		workspaces:     config.Workspaces,
		workflow:       config.Workflow,
		dispatcher:     config.Dispatcher,
		interval:       config.Interval,
		maxActiveRuns:  config.MaxActiveRuns,
		orchestratorID: config.OrchestratorID,
		refreshCh:      make(chan struct{}, 1),
	}, nil
}

func (p *Poller) Start(ctx context.Context) {
	go p.loop(ctx)
}

func (p *Poller) RequestRefresh() {
	select {
	case p.refreshCh <- struct{}{}:
	default:
	}
}

func (p *Poller) Poll(ctx context.Context) error {
	queue, err := p.tracker.Queue(ctx)
	if err != nil {
		return fmt.Errorf("poll issue tracker: %w", err)
	}
	activeRuns, err := p.store.ActiveRuns(ctx)
	if err != nil {
		return fmt.Errorf("list active runs: %w", err)
	}
	activeIssueIDs := make(map[int64]struct{}, len(activeRuns))
	for _, storedRun := range activeRuns {
		activeIssueIDs[storedRun.IssueID] = struct{}{}
	}

	created := 0
	availableSlots := p.maxActiveRuns - len(activeRuns)
	if availableSlots > 0 {
		for _, queuedIssue := range queue.Queued {
			issue := queuedIssue.Issue
			if _, ok := activeIssueIDs[issue.ID]; ok {
				continue
			}
			if created >= availableSlots {
				break
			}
			if _, err := p.queueIssue(ctx, issue); err != nil {
				return err
			}
			created++
		}
	}
	if p.dispatcher != nil {
		activeRuns, err = p.store.ActiveRuns(ctx)
		if err != nil {
			return fmt.Errorf("list active runs for dispatch: %w", err)
		}
		if err := p.dispatcher.Dispatch(ctx, activeRuns); err != nil {
			return fmt.Errorf("dispatch active runs: %w", err)
		}
	}
	log.Printf("orchestrator poll complete queued_available=%d pending=%d queued=%d active=%d max_active=%d", len(queue.Queued), len(queue.Pending), created, len(activeRuns), p.maxActiveRuns)
	return nil
}

func (p *Poller) loop(ctx context.Context) {
	p.runPoll(ctx, "startup")
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.runPoll(ctx, "interval")
		case <-p.refreshCh:
			p.runPoll(ctx, "refresh")
		}
	}
}

func (p *Poller) runPoll(ctx context.Context, reason string) {
	if err := p.Poll(ctx); err != nil && ctx.Err() == nil {
		log.Printf("orchestrator poll failed reason=%s: %v", reason, err)
	}
}

func (p *Poller) queueIssue(ctx context.Context, issue entity.Issue) (run.Run, error) {
	identifier := issueIdentifier(issue.ID)
	project, err := p.tracker.Project(ctx, issue.ProjectID)
	if err != nil {
		return run.Run{}, fmt.Errorf("fetch project %d for %s: %w", issue.ProjectID, identifier, err)
	}
	definition, err := p.workflow.Resolve(ctx, project.ID)
	if err != nil {
		_ = p.store.RecordWorkspaceSetupFailure(ctx, issue.ID, identifier, "", err.Error())
		return run.Run{}, fmt.Errorf("resolve workflow for project %d: %w", project.ID, err)
	}
	workspaceRoot, err := projectWorkspaceRoot(definition, project.Location)
	if err != nil {
		_ = p.store.RecordWorkspaceSetupFailure(ctx, issue.ID, identifier, "", err.Error())
		return run.Run{}, fmt.Errorf("resolve workspace root for %s: %w", identifier, err)
	}
	workspaceInfo, err := p.workspaces.CreateForIssueInProjectLocationWithConfig(identifier, project.Location, workspaceRoot, hookConfig(definition.Config))
	if err != nil {
		_ = p.store.RecordWorkspaceSetupFailure(ctx, issue.ID, identifier, "", err.Error())
		return run.Run{}, fmt.Errorf("create workspace for %s: %w", identifier, err)
	}
	if err := p.store.UpsertWorkspaceMetadata(ctx, runstore.WorkspaceMetadataInput{
		WorkspaceKey: workspaceInfo.WorkspaceKey,
		IssueID:      issue.ID,
		Path:         workspaceInfo.Path,
		CreatedNow:   workspaceInfo.CreatedNow,
		SourcePath:   "",
	}); err != nil {
		return run.Run{}, fmt.Errorf("record workspace metadata for %s: %w", identifier, err)
	}
	attempt, err := p.nextAttempt(ctx, issue.ID)
	if err != nil {
		return run.Run{}, err
	}
	createdRun, err := p.store.CreateRun(ctx, runstore.CreateRunInput{
		IssueID:        issue.ID,
		Workspace:      workspaceInfo.Path,
		Attempt:        attempt,
		OrchestratorID: p.orchestratorID,
	})
	if err != nil {
		return run.Run{}, fmt.Errorf("create run for %s: %w", identifier, err)
	}
	if err := p.store.RecordRunnerEvent(ctx, createdRun.RunID, "queued", "issue queued from tracker poll", ""); err != nil {
		return run.Run{}, fmt.Errorf("record queued event for %s: %w", identifier, err)
	}
	return createdRun, nil
}

func (p *Poller) nextAttempt(ctx context.Context, issueID int64) (int, error) {
	latestRun, err := p.store.RunByIssueID(ctx, issueID)
	if errors.Is(err, sql.ErrNoRows) {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read latest run for issue %d: %w", issueID, err)
	}
	return latestRun.Attempt + 1, nil
}

func issueIdentifier(issueID int64) string {
	return fmt.Sprintf("issue-%d", issueID)
}

func projectWorkspaceRoot(definition workflow.Definition, projectLocation string) (string, error) {
	root := definition.Config.WorkspaceRoot
	if strings.TrimSpace(projectLocation) == "" || strings.TrimSpace(definition.Path) == "" {
		return root, nil
	}
	rel, err := filepath.Rel(filepath.Dir(definition.Path), root)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return root, nil
	}
	return filepath.Join(projectLocation, rel), nil
}

func hookConfig(config workflow.Config) workspace.HookConfig {
	return workspace.HookConfig{
		AfterCreate:  config.HookAfterCreate,
		BeforeRun:    config.HookBeforeRun,
		AfterRun:     config.HookAfterRun,
		BeforeRemove: config.HookBeforeRemove,
		Timeout:      config.HookTimeout,
	}
}
