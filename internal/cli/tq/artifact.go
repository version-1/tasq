package tq

import (
	"context"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (a app) routeArtifact(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printArtifactHelp(a.stdout)
		return nil
	}
	switch args[0] {
	case "document":
		return a.routeArtifactDocument(ctx, args[1:], cfg)
	case "set":
		return a.artifactSet(ctx, args[1:], cfg)
	case "delete":
		return a.artifactDelete(ctx, args[1:], cfg)
	default:
		return usageError("unknown artifact action %q", args[0])
	}
}

func (a app) routeArtifactDocument(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 {
		return usageError("usage: tq artifact document <create|list|get|update> ...")
	}
	switch args[0] {
	case "create":
		return a.createDocumentArtifact(ctx, args[1:], cfg)
	case "list":
		return a.listDocumentArtifacts(ctx, args[1:], cfg)
	case "get":
		return a.getDocumentArtifact(ctx, args[1:], cfg)
	case "update":
		return a.updateDocumentArtifact(ctx, args[1:], cfg)
	default:
		return usageError("unknown document artifact action %q", args[0])
	}
}
func (a app) createDocumentArtifact(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("artifact document create")
	artifactType := fs.String("type", "", "reference_document or work_report")
	title := fs.String("title", "", "document title")
	file := fs.String("file", "", "Markdown file path or - for standard input")
	if len(args) == 0 {
		return usageError("usage: tq artifact document create <issue-id> --type TYPE --title TITLE --file PATH|-")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	if err := fs.Parse(args[1:]); err != nil {
		return usageError("%s", err)
	}
	if fs.NArg() != 0 || *artifactType == "" || *title == "" || *file == "" {
		return usageError("usage: tq artifact document create <issue-id> --type TYPE --title TITLE --file PATH|-")
	}
	body, err := readDocumentArtifactBody(*file, a.stdin)
	if err != nil {
		return err
	}
	item, err := a.client.createDocumentArtifact(ctx, id, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactType(*artifactType), Title: *title, Body: body})
	if err != nil {
		return err
	}
	return writeDocumentArtifact(a.stdout, cfg.output, item)
}
func (a app) listDocumentArtifacts(ctx context.Context, args []string, cfg config) error {
	if len(args) != 1 {
		return usageError("usage: tq artifact document list <issue-id>")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	items, err := a.client.listDocumentArtifacts(ctx, id)
	if err != nil {
		return err
	}
	return writeDocumentArtifacts(a.stdout, cfg.output, items)
}
func (a app) getDocumentArtifact(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("artifact document get")
	raw := fs.Bool("raw", false, "write only Markdown body")
	if len(args) < 2 {
		return usageError("usage: tq artifact document get <issue-id> <document-id>")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	if err := fs.Parse(args[2:]); err != nil || fs.NArg() != 0 {
		return usageError("usage: tq artifact document get <issue-id> <document-id> [--raw]")
	}
	item, err := a.client.documentArtifact(ctx, id, args[1])
	if err != nil {
		return err
	}
	if *raw {
		_, err := io.WriteString(a.stdout, item.Body)
		return err
	}
	return writeDocumentArtifact(a.stdout, cfg.output, item)
}
func (a app) updateDocumentArtifact(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("artifact document update")
	title := fs.String("title", "", "document title")
	file := fs.String("file", "", "Markdown file path or - for standard input")
	if len(args) < 2 {
		return usageError("usage: tq artifact document update <issue-id> <document-id> [--title TITLE] [--file PATH|-]")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	if err := fs.Parse(args[2:]); err != nil {
		return usageError("%s", err)
	}
	if fs.NArg() != 0 || (*title == "" && *file == "") {
		return usageError("usage: tq artifact document update <issue-id> <document-id> [--title TITLE] [--file PATH|-]")
	}
	input := entity.UpdateDocumentArtifactInput{}
	if *title != "" {
		input.Title = title
	}
	if *file != "" {
		body, err := readDocumentArtifactBody(*file, a.stdin)
		if err != nil {
			return err
		}
		input.Body = &body
	}
	item, err := a.client.updateDocumentArtifact(ctx, id, args[1], input)
	if err != nil {
		return err
	}
	return writeDocumentArtifact(a.stdout, cfg.output, item)
}
func readDocumentArtifactBody(path string, stdin io.Reader) (string, error) {
	const maxDocumentBytes = 100000
	if path == "-" {
		bytes, err := io.ReadAll(io.LimitReader(stdin, maxDocumentBytes+1))
		if err != nil {
			return "", fmt.Errorf("read document standard input: %w", err)
		}
		if len(bytes) > maxDocumentBytes {
			return "", usageError("document body must be 100000 bytes or fewer")
		}
		if !utf8.Valid(bytes) {
			return "", usageError("document body is invalid UTF-8")
		}
		return string(bytes), nil
	}
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read document file: %w", err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(io.LimitReader(file, maxDocumentBytes+1))
	if err != nil {
		return "", fmt.Errorf("read document file: %w", err)
	}
	if len(bytes) > maxDocumentBytes {
		return "", usageError("document body must be 100000 bytes or fewer")
	}
	if !utf8.Valid(bytes) {
		return "", usageError("document body is invalid UTF-8")
	}
	return string(bytes), nil
}

func (a app) artifactSet(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("artifact set")
	artifactType := fs.String("type", "", "artifact type")
	if len(args) == 0 {
		return usageError("usage: tq artifact set <issue_id> --type pull_request <url>")
	}
	issueID, err := parseID(args[0])
	if err != nil {
		return err
	}
	if err := fs.Parse(args[1:]); err != nil {
		return usageError("%s", err)
	}
	if fs.NArg() != 1 {
		return usageError("usage: tq artifact set <issue_id> --type pull_request <url>")
	}
	if *artifactType == "" {
		return usageError("type is required")
	}
	artifact, err := a.client.upsertArtifact(ctx, issueID, *artifactType, fs.Arg(0))
	if err != nil {
		return err
	}
	return writeArtifact(a.stdout, cfg.output, artifact)
}

func (a app) artifactDelete(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("artifact delete")
	artifactType := fs.String("type", "", "artifact type")
	if len(args) == 0 {
		return usageError("usage: tq artifact delete <issue_id> --type pull_request")
	}
	issueID, err := parseID(args[0])
	if err != nil {
		return err
	}
	if err := fs.Parse(args[1:]); err != nil {
		return usageError("%s", err)
	}
	if fs.NArg() != 0 {
		return usageError("usage: tq artifact delete <issue_id> --type pull_request")
	}
	if *artifactType == "" {
		return usageError("type is required")
	}
	if err := a.client.deleteArtifact(ctx, issueID, *artifactType); err != nil {
		return err
	}
	return writeArtifactDeleted(a.stdout, cfg.output, issueID, *artifactType)
}
