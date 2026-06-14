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
