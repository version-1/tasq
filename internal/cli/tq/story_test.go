package tq

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunStoryRendersEveryScenario(t *testing.T) {
	for _, story := range outputStories() {
		t.Run(story.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := RunStory([]string{story.name}, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			if stdout.Len() == 0 {
				t.Fatal("stdout is empty")
			}
		})
	}
}

func TestRunStoryUnknownScenarioShowsUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := RunStory([]string{"missing"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "tq_service_start_fail") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestRunStoryJSONScenariosContainNoANSI(t *testing.T) {
	for _, name := range []string{"tq_json_success", "tq_json_error"} {
		t.Run(name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := RunStory([]string{name}, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			if strings.Contains(stdout.String(), "\x1b[") {
				t.Fatalf("story contains ANSI: %q", stdout.String())
			}
		})
	}
}

func TestRunStoryCommandResultsUseSharedRenderers(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "tq_service_start_success", want: "Services started"},
		{name: "tq_service_stop_success", want: "Services stopped"},
		{name: "tq_project_remove_success", want: "Removed project"},
		{name: "tq_workflow_remove_success", want: "Removed workflow override"},
		{name: "tq_project_remove_cancelled", want: "Project removal cancelled"},
		{name: "tq_project_remove_confirmation", want: "Type the project key to confirm"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			if code := RunStory([]string{test.name}, &stdout, &stderr); code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), test.want) {
				t.Fatalf("output does not contain %q: %s", test.want, stdout.String())
			}
		})
	}
}
