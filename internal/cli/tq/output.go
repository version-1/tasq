package tq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

const (
	ansiBold    = "\x1b[1m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiRed     = "\x1b[31m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
	ansiFaint   = "\x1b[2m"
	ansiReset   = "\x1b[0m"
)

func writeIssues(w io.Writer, format string, issues []entity.Issue) error {
	if format == "json" {
		return writeJSON(w, issues)
	}
	for _, issue := range issues {
		fmt.Fprintf(w, "%s#%d%s\t[%s%s%s]\t%s\t%s\n", ansiBold, issue.ID, ansiReset, ansiCyan, issue.ProjectKey, ansiReset, colorValue(string(issue.Status), statusColor(issue.Status)), issue.Title)
	}
	return nil
}

func writeIssue(w io.Writer, format string, issue entity.Issue) error {
	if format == "json" {
		return writeJSON(w, issue)
	}
	fmt.Fprintf(w, "ID: %d\n", issue.ID)
	fmt.Fprintf(w, "Project: %s\n", issue.ProjectKey)
	fmt.Fprintf(w, "Title: %s\n", issue.Title)
	fmt.Fprintf(w, "Description: %s\n", issue.Description)
	fmt.Fprintf(w, "Status: %s\n", colorValue(string(issue.Status), statusColor(issue.Status)))
	fmt.Fprintf(w, "Priority: %s\n", colorValue(string(issue.Priority), priorityColor(issue.Priority)))
	fmt.Fprintf(w, "Assignee: %s\n", valueOrDefault(issue.Assignee, "unassigned"))
	fmt.Fprintf(w, "Dependencies: %s\n", formatDependencyIDs(issue.DependencyIDs))
	fmt.Fprintf(w, "Created: %s\n", formatTime(issue.CreatedAt))
	fmt.Fprintf(w, "Updated: %s\n", formatTime(issue.UpdatedAt))
	return nil
}

func writeIssueAction(w io.Writer, format string, issue entity.Issue, message string) error {
	if format == "json" {
		return writeIssue(w, format, issue)
	}
	if _, err := fmt.Fprintf(w, "%s✓%s %s\n", ansiGreen, ansiReset, message); err != nil {
		return err
	}
	return writeIssue(w, format, issue)
}

func writeSuccess(w io.Writer, message string) error {
	_, err := fmt.Fprintf(w, "%s✓%s %s\n", ansiGreen, ansiReset, message)
	return err
}

func writeFaintMessage(w io.Writer, message string) error {
	_, err := fmt.Fprintf(w, "%s%s%s\n", ansiFaint, message, ansiReset)
	return err
}

func writeProjectRemovalConfirmation(w io.Writer, project entity.Project) error {
	if _, err := fmt.Fprintf(w, "%sWARNING:%s This operation cannot be undone.\n", ansiRed, ansiReset); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Project to remove: %s (%s)\n", project.Key, project.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "This deletes the project and descendant data, including issues, comments, attachments, workflow overrides, and run data."); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, "Type the project key to confirm: ")
	return err
}

func writeProjectRemoved(w io.Writer, projectKey string) error {
	return writeSuccess(w, "Removed project "+colorValue(projectKey, ansiCyan))
}

func writeWorkflowOverrideRemoved(w io.Writer, projectKey string) error {
	return writeSuccess(w, "Removed workflow override for project "+colorValue(projectKey, ansiCyan))
}

func writeServicesStarted(w io.Writer) error {
	return writeSuccess(w, "Services started")
}

func writeServicesStopped(w io.Writer) error {
	return writeSuccess(w, "Services stopped")
}

func colorValue(value string, color string) string {
	if color == "" {
		return value
	}
	return color + value + ansiReset
}

func statusColor(status entity.Status) string {
	switch status {
	case entity.StatusBacklog:
		return ansiFaint
	case entity.StatusReady:
		return ansiCyan
	case entity.StatusInProgress:
		return ansiYellow
	case entity.StatusReview:
		return ansiMagenta
	case entity.StatusBlocked, entity.StatusFailed:
		return ansiRed
	case entity.StatusCancelled, entity.StatusDuplicate:
		return ansiFaint
	case entity.StatusDone:
		return ansiGreen
	default:
		return ""
	}
}

func priorityColor(priority entity.Priority) string {
	switch priority {
	case entity.PriorityLow:
		return ansiFaint
	case entity.PriorityHigh:
		return ansiYellow
	case entity.PriorityUrgent:
		return ansiRed
	default:
		return ""
	}
}

func writeProjects(w io.Writer, format string, projects []entity.Project) error {
	if format == "json" {
		return writeJSON(w, projects)
	}
	if len(projects) == 0 {
		_, err := fmt.Fprintf(w, "%sNo projects found.%s\n", ansiFaint, ansiReset)
		return err
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tKEY\tNAME\tLOCATION\tUPDATED")
	for _, project := range projects {
		fmt.Fprintf(tw, "#%d\t%s\t%s\t%s\t%s\n", project.ID, project.Key, project.Name, project.Location, formatTime(project.UpdatedAt))
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := io.WriteString(w, colorProjectTable(buf.String()))
	return err
}

func colorProjectTable(table string) string {
	lines := strings.Split(table, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if i == 0 {
			lines[i] = ansiBold + line + ansiReset
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		lines[i] = strings.Replace(line, fields[1], ansiCyan+fields[1]+ansiReset, 1)
	}
	return strings.Join(lines, "\n")
}

func writeProjectAddResult(w io.Writer, format string, result projectAddResult) error {
	if format == "json" {
		return writeJSON(w, result)
	}
	fmt.Fprintf(w, "Project: %s (%d)\n", result.Project.Key, result.Project.ID)
	fmt.Fprintf(w, "Path: %s\n", result.Project.Location)
	return nil
}

func writeProjectCheckItems(w io.Writer, format string, items []projectCheckItem) error {
	if format == "json" {
		return writeJSON(w, items)
	}
	for _, item := range items {
		status := colorValue("PASS", ansiGreen)
		if !item.Passed {
			status = colorValue("FAIL", ansiRed)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", status, item.Name, item.Reason)
	}
	return nil
}

func writeServiceStatuses(w io.Writer, statuses []serviceStatus) error {
	for _, status := range statuses {
		if status.State == "running" {
			if _, err := fmt.Fprintf(w, "%s%s%s\t%s\tpid=%d\tport=%d\tuptime=%s\n", ansiCyan, status.Name, ansiReset, colorValue("running", ansiGreen), status.PID, status.Port, status.Uptime); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s%s%s\t%s\n", ansiCyan, status.Name, ansiReset, colorValue("stopped", ansiFaint)); err != nil {
			return err
		}
	}
	return nil
}

func writeWorkflowAddResult(w io.Writer, format string, project entity.Project, workflow entity.ProjectWorkflow) error {
	if format == "json" {
		return writeJSON(w, workflow)
	}
	fmt.Fprintf(w, "Workflow override updated for project %s\n", project.Key)
	fmt.Fprintf(w, "Checksum: %s\n", workflow.Checksum)
	return nil
}

func writeComments(w io.Writer, format string, comments []entity.Comment) error {
	if format == "json" {
		return writeJSON(w, comments)
	}
	for _, comment := range comments {
		fmt.Fprintf(w, "#%d\tissue:%d\t%s\t%s\t%s\n", comment.ID, comment.IssueID, comment.Type, comment.Author, comment.Body)
	}
	return nil
}

func writeComment(w io.Writer, format string, comment entity.Comment) error {
	if format == "json" {
		return writeJSON(w, comment)
	}
	fmt.Fprintf(w, "ID: %d\n", comment.ID)
	fmt.Fprintf(w, "Issue: %d\n", comment.IssueID)
	fmt.Fprintf(w, "Author: %s\n", comment.Author)
	fmt.Fprintf(w, "Type: %s\n", comment.Type)
	fmt.Fprintf(w, "Body: %s\n", comment.Body)
	fmt.Fprintf(w, "Created: %s\n", formatTime(comment.CreatedAt))
	return nil
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeCLIErrorForFormat(w io.Writer, format string, message string, code int) int {
	if format == "json" {
		_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
		return code
	}
	_, _ = fmt.Fprintf(w, "%sError:%s %s\n", ansiRed, ansiReset, message)
	return code
}

func writeMigrateResults(w io.Writer, format string, heading string, results []migrateResult) error {
	if format == "json" {
		return writeJSON(w, results)
	}
	fmt.Fprintln(w, ansiBold+heading+ansiReset)
	for _, result := range results {
		fmt.Fprintf(w, "%s%s%s\t%s\n", ansiCyan, result.Database, ansiReset, result.Path)
		switch {
		case result.Statuses != nil:
			for _, status := range result.Statuses {
				state := "pending"
				if status.Applied {
					state = "applied"
				}
				fmt.Fprintf(w, "  %s_%s\t%s\n", status.Version, status.Name, colorValue(state, migrationStateColor(status.Applied)))
			}
		case result.RolledBack != "":
			fmt.Fprintf(w, "  %srolled back%s %s\n", ansiYellow, ansiReset, result.RolledBack)
		case len(result.Applied) > 0:
			for _, item := range result.Applied {
				fmt.Fprintf(w, "  %sapplied%s %s\n", ansiGreen, ansiReset, item)
			}
		default:
			fmt.Fprintf(w, "  %sno changes%s\n", ansiFaint, ansiReset)
		}
	}
	return nil
}

func migrationStateColor(applied bool) string {
	if applied {
		return ansiGreen
	}
	return ansiYellow
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func formatDependencyIDs(ids []int64) string {
	if len(ids) == 0 {
		return "none"
	}
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, strconv.FormatInt(id, 10))
	}
	return strings.Join(values, ",")
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format(time.RFC3339)
}
