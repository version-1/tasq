package usecase

import (
	"context"
	"log"
	"time"

	"github.com/version-1/tasq/internal/orchestrator"
	"github.com/version-1/tasq/internal/orchestrator/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/infra/store"
)

type Worker struct {
	store          *store.Store
	client         *orchestrator.IssueTrackerClient
	orchestratorID string
	pollInterval   time.Duration
	leaseSeconds   int
}

func NewWorker(store *store.Store, client *orchestrator.IssueTrackerClient, orchestratorID string, pollInterval time.Duration, leaseSeconds int) *Worker {
	if pollInterval <= 0 {
		pollInterval = 3 * time.Second
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	return &Worker{
		store:          store,
		client:         client,
		orchestratorID: orchestratorID,
		pollInterval:   pollInterval,
		leaseSeconds:   leaseSeconds,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	item, err := w.client.ClaimWorkItem(ctx, w.orchestratorID, w.leaseSeconds)
	if err != nil {
		log.Printf("claim work item: %v", err)
		return
	}
	if item == nil {
		return
	}
	workspace := ".workspaces/issue-" + formatInt(item.IssueID)
	run, err := w.store.CreateRun(ctx, store.CreateRunInput{
		IssueID:        item.IssueID,
		WorkItemID:     item.ID,
		ClaimToken:     item.ClaimToken,
		Workspace:      workspace,
		Attempt:        item.Attempt,
		OrchestratorID: w.orchestratorID,
	})
	if err != nil {
		log.Printf("create run: %v", err)
		return
	}
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	if _, err := w.store.UpdateRunStatus(ctx, run.RunID, entity.RunRunning, ""); err != nil {
		log.Printf("mark run running: %v", err)
		return
	}
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	if _, err := w.store.UpdateRunStatus(ctx, run.RunID, entity.RunSucceeded, ""); err != nil {
		log.Printf("mark run succeeded: %v", err)
		return
	}
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
}

func (w *Worker) flushOutbox(ctx context.Context) error {
	events, err := w.store.UnsentOutboxEvents(ctx, 20)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := w.client.SendRunEvent(ctx, event); err != nil {
			return err
		}
		if err := w.store.MarkOutboxEventSent(ctx, event.ID); err != nil {
			return err
		}
	}
	return nil
}

func formatInt(value int64) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	negative := value < 0
	if negative {
		value = -value
	}
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
