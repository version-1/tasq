package runner

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func continuationPrompt(task Task) string {
	prompt := fmt.Sprintf(continuationGuidance, task.Issue.ID, task.Issue.ID)
	changeRequestGuidance := changeRequestGuidanceForTask(task)
	if changeRequestGuidance == "" {
		return prompt
	}
	return prompt + "\n\n" + changeRequestGuidance
}

func changeRequestGuidanceForTask(task Task) string {
	if len(task.ChangeRequests) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("Change requests assigned to this continuation:\n")
	for _, request := range task.ChangeRequests {
		builder.WriteString("- #")
		builder.WriteString(strconv.FormatInt(request.ID, 10))
		builder.WriteString(" by ")
		builder.WriteString(request.Author)
		builder.WriteString(": ")
		builder.WriteString(request.Body)
		builder.WriteString("\n")
	}
	builder.WriteString("\nAfter handling each change request, update it to `resolved` with `resolvedByRunId` set to this run ID")
	if task.RunID != "" {
		builder.WriteString(" (`")
		builder.WriteString(task.RunID)
		builder.WriteString("`)")
	}
	builder.WriteString(". Use `PATCH /api/v1/change-requests/{id}` on the issue-tracker API. Include `resultCommentId` when a result comment is available.")
	return builder.String()
}

func renderPrompt(task Task) (string, error) {
	prompt := task.PromptTemplate
	if prompt == "" {
		prompt = "Work on issue {{ issue.id }}: {{ issue.title }}\n\n{{ issue.description }}"
	}
	if shouldInjectTaskWorkPrompt(task.TaskWorkPrompt) {
		prompt = defaultTaskWorkPrompt + "\n\n" + prompt
	}
	if strings.Count(prompt, "{{") != strings.Count(prompt, "}}") {
		return "", errors.New("template_parse_error: unbalanced template delimiters")
	}
	var renderErr error
	vars := templateVariables(task)
	rendered := templateVariablePattern.ReplaceAllStringFunc(prompt, func(match string) string {
		if renderErr != nil {
			return match
		}
		parts := templateVariablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			renderErr = fmt.Errorf("template_parse_error: malformed template expression %q", match)
			return match
		}
		expression := strings.TrimSpace(parts[1])
		if strings.Contains(expression, "|") {
			name, _, _ := strings.Cut(expression, "|")
			if templateNamePattern.MatchString(strings.TrimSpace(name)) {
				renderErr = fmt.Errorf("template_render_error: unknown filter in %q", expression)
			}
			return match
		}
		if !templateNamePattern.MatchString(expression) {
			return match
		}
		value, ok := vars[expression]
		if !ok {
			renderErr = fmt.Errorf("template_render_error: unknown variable %q", expression)
			return match
		}
		return value
	})
	if renderErr != nil {
		return "", renderErr
	}
	if changeRequestGuidance := changeRequestGuidanceForTask(task); changeRequestGuidance != "" {
		rendered += "\n\n" + changeRequestGuidance
	}
	return rendered, nil
}

func shouldInjectTaskWorkPrompt(taskWorkPrompt *bool) bool {
	return taskWorkPrompt == nil || *taskWorkPrompt
}

func templateVariables(task Task) map[string]string {
	attempt := 0
	if task.Attempt > 1 {
		attempt = task.Attempt
	}
	return map[string]string{
		"issue.id":          strconv.FormatInt(task.Issue.ID, 10),
		"issue.title":       task.Issue.Title,
		"issue.description": task.Issue.Description,
		"issue.status":      string(task.Issue.Status),
		"issue.priority":    string(task.Issue.Priority),
		"issue.assignee":    task.Issue.Assignee,
		"issue.created_at":  task.Issue.CreatedAt.Format(time.RFC3339),
		"issue.updated_at":  task.Issue.UpdatedAt.Format(time.RFC3339),
		"attempt":           strconv.Itoa(attempt),
	}
}
