package entity

import "testing"

func TestIsValidStatusIncludesTerminalStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{name: "cancelled", status: StatusCancelled, want: true},
		{name: "duplicate", status: StatusDuplicate, want: true},
		{name: "unknown", status: Status("unknown"), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidStatus(test.status); got != test.want {
				t.Fatalf("IsValidStatus(%q) = %t, want %t", test.status, got, test.want)
			}
		})
	}
}

func TestOrderedStatusesIncludesTerminalStatuses(t *testing.T) {
	got := OrderedStatuses()
	want := []Status{
		StatusBacklog,
		StatusReady,
		StatusInProgress,
		StatusReview,
		StatusBlocked,
		StatusFailed,
		StatusCancelled,
		StatusDuplicate,
		StatusDone,
	}

	if len(got) != len(want) {
		t.Fatalf("len(OrderedStatuses()) = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("OrderedStatuses()[%d] = %q, want %q: %v", i, got[i], want[i], got)
		}
	}
}

func TestDependencyStatusClassification(t *testing.T) {
	active := map[Status]bool{
		StatusBacklog:    true,
		StatusReady:      true,
		StatusInProgress: true,
		StatusReview:     true,
	}
	satisfied := map[Status]bool{
		StatusDone:      true,
		StatusCancelled: true,
		StatusDuplicate: true,
	}

	for _, status := range OrderedStatuses() {
		gotActive := IsActiveDependencyStatus(status)
		gotSatisfied := IsSatisfiedDependencyStatus(status)
		if gotActive != active[status] {
			t.Fatalf("IsActiveDependencyStatus(%q) = %t, want %t", status, gotActive, active[status])
		}
		if gotSatisfied != satisfied[status] {
			t.Fatalf("IsSatisfiedDependencyStatus(%q) = %t, want %t", status, gotSatisfied, satisfied[status])
		}
		if gotActive && gotSatisfied {
			t.Fatalf("dependency status %q must not be both active and satisfied", status)
		}
	}
	if IsActiveDependencyStatus(Status("unknown")) {
		t.Fatal("unknown status must not be active")
	}
	if IsSatisfiedDependencyStatus(Status("unknown")) {
		t.Fatal("unknown status must not be satisfied")
	}
}

func TestIssueQueueStatus(t *testing.T) {
	tests := []struct {
		name                string
		status              Status
		hasActiveDependency bool
		want                QueueStatus
	}{
		{name: "backlog", status: StatusBacklog, want: QueueStatusBacklog},
		{name: "ready without active dependency", status: StatusReady, want: QueueStatusQueued},
		{name: "ready with active dependency", status: StatusReady, hasActiveDependency: true, want: QueueStatusPending},
		{name: "in progress", status: StatusInProgress, want: QueueStatusProcessing},
		{name: "review", status: StatusReview, want: QueueStatusInactive},
		{name: "done", status: StatusDone, want: QueueStatusCompleted},
		{name: "blocked", status: StatusBlocked, want: QueueStatusInactive},
		{name: "failed", status: StatusFailed, want: QueueStatusInactive},
		{name: "cancelled", status: StatusCancelled, want: QueueStatusInactive},
		{name: "duplicate", status: StatusDuplicate, want: QueueStatusInactive},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IssueQueueStatus(test.status, test.hasActiveDependency); got != test.want {
				t.Fatalf("IssueQueueStatus(%q, %t) = %q, want %q", test.status, test.hasActiveDependency, got, test.want)
			}
		})
	}
}
