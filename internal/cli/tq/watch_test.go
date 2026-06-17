package tq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestListIssuesByStatesSendsStatesQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("states"); got != "ready" {
			t.Fatalf("states query=%q, want %q", got, "ready")
		}
		writeTestJSON(t, w, apiResponse[[]entity.Issue]{
			Data: []entity.Issue{{ID: 1, Status: entity.StatusReady}},
		})
	}))
	defer server.Close()

	client, err := newAPIClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	issues, err := client.listIssuesByStates(context.Background(), []entity.Status{entity.StatusReady})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].ID != 1 || issues[0].Status != entity.StatusReady {
		t.Fatalf("unexpected issues: %+v", issues)
	}
}

type testEnvelope struct {
	Type      string          `json:"type"`
	EventType string          `json:"eventType"`
	Body      json.RawMessage `json:"body"`
}

func parseEnvelopes(t *testing.T, output string) []testEnvelope {
	t.Helper()
	var envelopes []testEnvelope
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var env testEnvelope
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			t.Fatalf("invalid envelope line %q: %v", line, err)
		}
		envelopes = append(envelopes, env)
	}
	return envelopes
}

func TestSeenSetTryMarkTTL(t *testing.T) {
	now := time.Unix(0, 0)
	set := newSeenSet(40*time.Second, func() time.Time { return now })

	if !set.tryMark(1) {
		t.Fatal("first sighting should be emittable")
	}
	now = now.Add(30 * time.Second)
	if set.tryMark(1) {
		t.Fatal("within TTL should be suppressed")
	}
	now = now.Add(20 * time.Second) // total 50s since first emit
	if !set.tryMark(1) {
		t.Fatal("after TTL should be emittable again")
	}
	now = now.Add(10 * time.Second) // 10s since the re-emit
	if set.tryMark(1) {
		t.Fatal("TTL should restart from the most recent emit")
	}
}

func TestSeenSetPruneBoundsMemory(t *testing.T) {
	now := time.Unix(0, 0)
	set := newSeenSet(40*time.Second, func() time.Time { return now })
	set.tryMark(1)
	now = now.Add(40 * time.Second)
	set.prune()
	if _, ok := set.seen[1]; ok {
		t.Fatal("expired entry should be pruned")
	}
}

func TestWatcherDedupAndTTLReEmit(t *testing.T) {
	var buf bytes.Buffer
	now := time.Unix(0, 0)
	cycle := 0
	w := &watcher{
		fetch: func(context.Context) ([]entity.Issue, error) {
			return []entity.Issue{{ID: 1, Status: entity.StatusReady}}, nil
		},
		interval: 30 * time.Second,
		seen:     newSeenSet(40*time.Second, func() time.Time { return now }),
		out:      &buf,
		sleep: func(context.Context, time.Duration) error {
			cycle++
			now = now.Add(30 * time.Second)
			if cycle >= 3 {
				return context.Canceled
			}
			return nil
		},
	}
	if err := w.run(context.Background()); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	envelopes := parseEnvelopes(t, buf.String())
	events := 0
	for _, env := range envelopes {
		if env.Type != "event" {
			t.Fatalf("unexpected envelope type without verbose: %q", env.Type)
		}
		if env.EventType != "issue-ready" {
			t.Fatalf("unexpected eventType: %q", env.EventType)
		}
		events++
	}
	// t=0 emit, t=30 suppressed (within 40s TTL), t=60 re-emit.
	if events != 2 {
		t.Fatalf("expected 2 events, got %d", events)
	}
}

func TestWatcherErrorEnvelopeContinuesLoop(t *testing.T) {
	var buf bytes.Buffer
	calls := 0
	cycle := 0
	w := &watcher{
		fetch: func(context.Context) ([]entity.Issue, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("api unreachable")
			}
			return []entity.Issue{{ID: 7, Status: entity.StatusReady}}, nil
		},
		interval: time.Second,
		seen:     newSeenSet(10*time.Second, func() time.Time { return time.Unix(0, 0) }),
		out:      &buf,
		sleep: func(context.Context, time.Duration) error {
			cycle++
			if cycle >= 2 {
				return context.Canceled
			}
			return nil
		},
	}
	if err := w.run(context.Background()); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	envelopes := parseEnvelopes(t, buf.String())
	if len(envelopes) != 2 {
		t.Fatalf("expected error then event, got %d envelopes: %s", len(envelopes), buf.String())
	}
	if envelopes[0].Type != "error" {
		t.Fatalf("first envelope should be error, got %q", envelopes[0].Type)
	}
	if envelopes[1].Type != "event" {
		t.Fatalf("loop should continue after error, got %q", envelopes[1].Type)
	}
}

func TestWatcherVerboseEmitsInfo(t *testing.T) {
	run := func(verbose bool) []testEnvelope {
		var buf bytes.Buffer
		w := &watcher{
			fetch: func(context.Context) ([]entity.Issue, error) {
				return []entity.Issue{{ID: 1, Status: entity.StatusReady}}, nil
			},
			interval: time.Second,
			seen:     newSeenSet(10*time.Second, func() time.Time { return time.Unix(0, 0) }),
			verbose:  verbose,
			out:      &buf,
			sleep: func(context.Context, time.Duration) error {
				return context.Canceled
			},
		}
		if err := w.run(context.Background()); err != nil {
			t.Fatalf("run returned error: %v", err)
		}
		return parseEnvelopes(t, buf.String())
	}

	for _, env := range run(false) {
		if env.Type == "info" {
			t.Fatal("info should not be emitted without verbose")
		}
	}

	infos := 0
	for _, env := range run(true) {
		if env.Type == "info" {
			infos++
		}
	}
	if infos == 0 {
		t.Fatal("verbose should emit info envelopes")
	}
}

func TestWatcherEventBodyIsFullIssue(t *testing.T) {
	var buf bytes.Buffer
	issue := entity.Issue{ID: 42, ProjectKey: "TQ", Title: "Add watch", Status: entity.StatusReady, Priority: entity.PriorityHigh}
	w := &watcher{
		fetch: func(context.Context) ([]entity.Issue, error) {
			return []entity.Issue{issue}, nil
		},
		interval: time.Second,
		seen:     newSeenSet(10*time.Second, func() time.Time { return time.Unix(0, 0) }),
		out:      &buf,
		sleep: func(context.Context, time.Duration) error {
			return context.Canceled
		},
	}
	if err := w.run(context.Background()); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	envelopes := parseEnvelopes(t, buf.String())
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 event, got %d", len(envelopes))
	}
	var got entity.Issue
	if err := json.Unmarshal(envelopes[0].Body, &got); err != nil {
		t.Fatalf("event body is not a full issue: %v", err)
	}
	if got != issue {
		t.Fatalf("event body=%+v, want %+v", got, issue)
	}
}

func TestIssueWatchRejectsInvalidFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "non-positive interval", args: []string{"issue", "watch", "--interval", "0"}},
		{name: "seen-ttl not greater than interval", args: []string{"issue", "watch", "--interval", "30", "--seen-ttl", "30"}},
		{name: "positional argument", args: []string{"issue", "watch", "extra"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := append([]string{"--api-url", "http://localhost:9999"}, test.args...)
			_, _, code := runCLI(t, args)
			if code != 2 {
				t.Fatalf("expected usage error exit code 2, got %d", code)
			}
		})
	}
}
