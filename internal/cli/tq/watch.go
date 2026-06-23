package tq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

// envelope is the single-line JSON protocol that `tq issue watch` writes to
// stdout. Each line is one notification for the Monitor tool: one of "error",
// "info", or "event". The global --output flag is ignored; watch always emits
// this envelope.
type envelope struct {
	Type      string `json:"type"`
	EventType string `json:"eventType,omitempty"`
	Body      any    `json:"body"`
}

// seenSet tracks which issue ids have already been emitted so the watch loop
// does not re-dispatch the same queued issue every cycle. An entry expires after
// ttl (measured from its last emit), after which the issue may be emitted again.
type seenSet struct {
	ttl  time.Duration
	now  func() time.Time
	seen map[int64]time.Time
}

func newSeenSet(ttl time.Duration, now func() time.Time) *seenSet {
	return &seenSet{ttl: ttl, now: now, seen: make(map[int64]time.Time)}
}

// tryMark reports whether id should be emitted now, recording the emit time
// when it returns true. An id is emittable when it has not been seen before or
// its previous emit is older than the TTL.
func (s *seenSet) tryMark(id int64) bool {
	now := s.now()
	if last, ok := s.seen[id]; ok && now.Sub(last) < s.ttl {
		return false
	}
	s.seen[id] = now
	return true
}

// prune drops entries whose TTL has elapsed to bound memory growth over a
// long-running watch. A pruned id that is still ready is treated as new on the
// next cycle, which is the same outcome as a TTL-expiry re-emit.
func (s *seenSet) prune() {
	now := s.now()
	for id, last := range s.seen {
		if now.Sub(last) >= s.ttl {
			delete(s.seen, id)
		}
	}
}

type issueFetcher func(ctx context.Context) ([]entity.Issue, error)

type watcher struct {
	fetch    issueFetcher
	interval time.Duration
	seen     *seenSet
	verbose  bool
	out      io.Writer
	sleep    func(ctx context.Context, d time.Duration) error
}

// run polls for queued issues until the context is cancelled. Transient fetch
// failures emit an error envelope and the loop continues; only context
// cancellation stops the loop, which it treats as a clean shutdown.
func (w *watcher) run(ctx context.Context) error {
	w.info(fmt.Sprintf("polling started (interval=%s, seen-ttl=%s, source=queue.queued)", w.interval, w.seen.ttl))
	for {
		issues, err := w.fetch(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.emitError(err.Error())
		} else {
			emitted := 0
			for _, issue := range issues {
				if w.seen.tryMark(issue.ID) {
					w.emitEvent(issue)
					emitted++
				}
			}
			w.seen.prune()
			w.info(fmt.Sprintf("polled: queued=%d, emitted=%d, skipped(ttl)=%d", len(issues), emitted, len(issues)-emitted))
		}
		if err := w.sleep(ctx, w.interval); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}
	}
}

func (w *watcher) info(message string) {
	if !w.verbose {
		return
	}
	w.write(envelope{Type: "info", Body: message})
}

func (w *watcher) emitError(message string) {
	w.write(envelope{Type: "error", Body: message})
}

func (w *watcher) emitEvent(issue entity.Issue) {
	w.write(envelope{Type: "event", EventType: "issue-ready", Body: issue})
}

func (w *watcher) write(env envelope) {
	enc := json.NewEncoder(w.out)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(env)
}

// sleepCtx waits for d or until the context is cancelled, whichever comes first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (a app) issueWatch(ctx context.Context, args []string) error {
	fs := newFlagSet("issue watch")
	interval := fs.Int("interval", 30, "polling interval in seconds")
	seenTTL := fs.Int("seen-ttl", 900, "seconds to suppress re-emitting the same issue")
	verbose := fs.Bool("verbose", false, "also emit info envelopes (error and event are always emitted)")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("issue watch does not accept positional arguments")
	}
	if *interval <= 0 {
		return usageError("interval must be a positive number of seconds")
	}
	if *seenTTL <= *interval {
		return usageError("seen-ttl must be greater than interval")
	}

	w := &watcher{
		fetch: func(ctx context.Context) ([]entity.Issue, error) {
			queue, err := a.client.issueQueue(ctx)
			if err != nil {
				return nil, err
			}
			issues := make([]entity.Issue, 0, len(queue.Queued))
			for _, issue := range queue.Queued {
				issues = append(issues, issue.Issue)
			}
			return issues, nil
		},
		interval: time.Duration(*interval) * time.Second,
		seen:     newSeenSet(time.Duration(*seenTTL)*time.Second, time.Now),
		verbose:  *verbose,
		out:      a.stdout,
		sleep:    sleepCtx,
	}
	return w.run(ctx)
}
