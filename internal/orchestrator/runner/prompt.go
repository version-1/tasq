package runner

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const approvalRequestGuidance = "Before requesting approval for a command execution or file change, provide a non-empty, specific reason. The reason must identify what needs approval, the target scope (the command and working directory or the file paths), why approval is required, and the expected effect. Do not send a null, empty, or vague reason such as only saying that approval is required."

const defaultTaskWorkPrompt = "Use `{{ tq.command }}` to keep the issue tracker synchronized. If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.\n" +
	"\n" +
	"Tracker access:\n" +
	"- Use `{{ tq.command }}` for every Tasq CLI operation, including commands shown elsewhere as `tq`. Do not substitute another executable from `PATH` or use `go run ./cmd/tq`.\n" +
	"- Prefer typed commands such as `{{ tq.command }} issue`, `{{ tq.command }} comment`, and `{{ tq.command }} artifact`. Use `{{ tq.command }} api` only when no typed command supports the operation.\n" +
	"- Do not call the issue tracker API directly with `curl`, `wget`, or a custom HTTP script. This restriction does not apply to other services or local endpoint verification.\n" +
	"\n" +
	"Lifecycle:\n" +
	"- At start, move the issue to `in_progress` and leave a progress comment. Add more progress comments at meaningful milestones.\n" +
	"- " + approvalRequestGuidance + "\n" +
	"- If blocked, leave a blocker comment that explains the blocker and what is needed, then move the issue to `blocked`.\n" +
	"- If you create or update a pull request, register the primary PR as the issue's `pull_request` artifact before handoff. Setting the artifact again replaces its URL; mention supporting PRs only in the handoff comment. If registration fails, retry reasonably; if it remains unresolved, leave a blocker comment and do not move to `review`. Skip registration when no pull request was created or updated.\n" +
	"- After any required artifact registration succeeds, leave a handoff comment with the PR URL and verification summary, then move the issue to `review`.\n" +
	"- Always pass `--author codex` when posting comments.\n" +
	"- Run only the commands for the current lifecycle stage; the examples below are alternatives, not a single script.\n" +
	"\n" +
	"```sh\n" +
	"# Start\n" +
	"{{ tq.command }} issue update {{ issue.id }} --status in_progress\n" +
	"{{ tq.command }} comment add {{ issue.id }} --author codex --type progress --body \"Started work.\"\n" +
	"\n" +
	"# Meaningful progress milestone\n" +
	"{{ tq.command }} comment add {{ issue.id }} --author codex --type progress --body \"Implemented the change; running verification.\"\n" +
	"\n" +
	"# Blocked (use instead of the review handoff)\n" +
	"{{ tq.command }} comment add {{ issue.id }} --author codex --type blocker --body \"Blocked: explain the blocker and what is needed.\"\n" +
	"{{ tq.command }} issue update {{ issue.id }} --status blocked\n" +
	"\n" +
	"# Ready for review (include the artifact command only when a PR was created or updated)\n" +
	"{{ tq.command }} artifact set {{ issue.id }} --type pull_request <pr-url>\n" +
	"{{ tq.command }} comment add {{ issue.id }} --author codex --type handoff --body \"PR: <url>; verification: <summary>.\"\n" +
	"{{ tq.command }} issue update {{ issue.id }} --status review\n" +
	"```"

const continuationGuidance = "First run `%[1]s issue update %[2]d --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread without repeating completed work, and stop when it is ready for handoff. " + approvalRequestGuidance + " If this continuation creates or updates a pull request, register the primary PR before handoff with `%[1]s artifact set %[2]d --type pull_request <pr-url>`. On success, add the handoff comment, then move the issue to `review`; on failure, retry reasonably, then leave a blocker comment and do not move to `review` if it remains unresolved. Otherwise, artifact registration is not required."

func continuationPrompt(task Task) string {
	prompt := fmt.Sprintf(continuationGuidance, tasqCommand(task), task.Issue.ID)
	if guidance := alternateTasqCommandGuidance(task); guidance != "" {
		prompt = guidance + "\n\n" + prompt
	}
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
		if guidance := alternateTasqCommandGuidance(task); guidance != "" {
			prompt = guidance + "\n\n" + prompt
		}
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
		"tq.command":        tasqCommand(task),
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

func tasqCommand(task Task) string {
	if task.TasqCommand != "" {
		return task.TasqCommand
	}
	return "tq"
}

func alternateTasqCommandGuidance(task Task) string {
	command := tasqCommand(task)
	if command == "tq" {
		return ""
	}
	return fmt.Sprintf("Use the `%[1]s` command instead of `tq`.\nWhen using the `tasq-cli` skill, interpret every `tq` command as `%[1]s`.", command)
}
