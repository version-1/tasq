package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func updateAutoOK(issue *entity.Issue, input entity.UpdateIssueInput) {
	if input.AutoOK != nil && *input.AutoOK != issue.AutoOK {
		issue.AutoOK = *input.AutoOK
		issue.AutoOKGeneration++
		if issue.AutoOK {
			issue.AutoOKCount = 0
			issue.AutoOKStopReason = ""
		} else if issue.AutoOKPending && input.Status == nil {
			issue.Status = entity.StatusBlocked
		}
		issue.AutoOKPending = false
	}
	// A manual status change cancels an unsent automatic reply.
	if input.Status != nil && *input.Status != entity.StatusReady && issue.AutoOKPending {
		issue.AutoOKPending = false
		issue.AutoOKGeneration++
	}
	if issue.Status == entity.StatusBlocked && issue.AutoOK && !issue.AutoOKPending && issue.AutoOKCount < entity.AutoOKLimit {
		issue.AutoOKGeneration++
		issue.AutoOKPending = true
		issue.Status = entity.StatusReady
	}
}

// CompleteAutoOK atomically records delivery (not task completion), its history,
// and the limit/failure stop. The generation prevents duplicate counting.
func (s *Store) CompleteAutoOK(ctx context.Context, id int64, input entity.AutoOKResultInput) (entity.Issue, error) {
	if input.Generation <= 0 {
		return entity.Issue{}, errors.New("generation must be positive")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Issue{}, err
	}
	defer tx.Rollback()
	issue, err := issueByIDTx(ctx, tx, id)
	if err != nil {
		return entity.Issue{}, err
	}
	if issue.AutoOKPending && issue.AutoOKGeneration == input.Generation {
		issue.AutoOKPending = false
		body := ""
		commentType := entity.CommentProgress
		if input.Error != "" {
			issue.AutoOK = false
			issue.AutoOKStopReason = "error"
			issue.Status = entity.StatusBlocked
			commentType = entity.CommentBlocker
			body = "Auto OK disabled: could not send ok and resume.\n\n" + input.Error
		} else {
			issue.AutoOKCount++
			if issue.Status == entity.StatusReady {
				issue.Status = entity.StatusInProgress
			}
			body = fmt.Sprintf("Auto OK: sent `ok` and resumed (%d/%d).", issue.AutoOKCount, entity.AutoOKLimit)
			if issue.AutoOKCount == entity.AutoOKLimit {
				issue.AutoOK = false
				issue.AutoOKStopReason = "limit"
				body += " Auto OK disabled after 5 replies."
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE issues SET status = ?, auto_ok = ?, auto_ok_count = ?, auto_ok_pending = 0, auto_ok_stop_reason = ?, updated_at = ? WHERE id = ?`, issue.Status, issue.AutoOK, issue.AutoOKCount, issue.AutoOKStopReason, nowString(), id)
		if err != nil {
			return entity.Issue{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO comments (issue_id, author, type, body, created_at) VALUES (?, ?, ?, ?, ?)`, id, "tasq-auto-ok", commentType, body, nowString())
		if err != nil {
			return entity.Issue{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return entity.Issue{}, err
	}
	return s.Issue(ctx, id)
}
