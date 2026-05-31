package tq

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func writeIssues(w io.Writer, format string, issues []entity.Issue) error {
	if format == "json" {
		return writeJSON(w, issues)
	}
	for _, issue := range issues {
		fmt.Fprintf(w, "#%d\t%s\t%s\n", issue.ID, issue.Status, issue.Title)
	}
	return nil
}

func writeIssue(w io.Writer, format string, issue entity.Issue) error {
	if format == "json" {
		return writeJSON(w, issue)
	}
	fmt.Fprintf(w, "ID: %d\n", issue.ID)
	fmt.Fprintf(w, "Title: %s\n", issue.Title)
	fmt.Fprintf(w, "Description: %s\n", issue.Description)
	fmt.Fprintf(w, "Status: %s\n", issue.Status)
	fmt.Fprintf(w, "Priority: %s\n", issue.Priority)
	fmt.Fprintf(w, "Assignee: %s\n", valueOrDefault(issue.Assignee, "unassigned"))
	fmt.Fprintf(w, "Created: %s\n", formatTime(issue.CreatedAt))
	fmt.Fprintf(w, "Updated: %s\n", formatTime(issue.UpdatedAt))
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

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format(time.RFC3339)
}
