package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

func TestAutoOKResumesWithOnlyOKAndAcknowledgesAcceptance(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "server.sh")
	if err := os.WriteFile(script, []byte(`#!/bin/sh
while IFS= read -r line; do
  case "$line" in
    *'"method":"initialize"'*) echo '{"id":1,"result":{}}' ;;
    *'"method":"thread/resume"'*) echo '{"id":2,"result":{"thread":{"id":"previous"}}}' ;;
    *'"method":"turn/start"'*)
      echo "$line" >> turns.jsonl
      echo '{"id":3,"result":{"turn":{"id":"turn-ok"}}}'
      echo '{"method":"turn/completed","params":{"threadId":"previous","turn":{"id":"turn-ok"}}}'
      ;;
  esac
done
`), 0700); err != nil {
		t.Fatal(err)
	}
	accepted := 0
	result := CodexRunner{}.Run(context.Background(), Task{
		AutoOK:            true,
		ResumeThreadID:    "previous",
		Workspace:         workspace.Workspace{Path: dir},
		Command:           "sh " + strconv.Quote(script),
		ReadTimeout:       time.Second,
		TurnTimeout:       time.Second,
		ContinueTurns:     true,
		MaxTurns:          5,
		PromptTemplate:    "Must not appear",
		OnAutoOKDelivered: func() error { accepted++; return nil },
	})
	if result.Status != run.StatusSucceeded || accepted != 1 {
		t.Fatalf("result=%+v accepted=%d", result, accepted)
	}
	data, err := os.ReadFile(filepath.Join(dir, "turns.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Params struct {
			Input []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"input"`
		} `json:"params"`
	}
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatalf("expected exactly one turn: %s: %v", data, err)
	}
	if len(request.Params.Input) != 1 || request.Params.Input[0].Text != "ok" || request.Params.Input[0].Type != "text" {
		t.Fatalf("input=%+v", request.Params.Input)
	}
}

func TestAutoOKRequiresExistingThread(t *testing.T) {
	result := CodexRunner{}.Run(context.Background(), Task{AutoOK: true, Command: "does-not-run", Workspace: workspace.Workspace{Path: t.TempDir()}})
	if result.Status != run.StatusFailed || result.Error != "Auto OK requires a previous thread to resume" {
		t.Fatalf("result=%+v", result)
	}
}
