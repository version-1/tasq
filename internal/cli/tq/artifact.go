package tq

import (
	"context"
)

func (a app) routeArtifact(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printArtifactHelp(a.stdout)
		return nil
	}
	switch args[0] {
	case "set":
		return a.artifactSet(ctx, args[1:], cfg)
	case "delete":
		return a.artifactDelete(ctx, args[1:], cfg)
	default:
		return usageError("unknown artifact action %q", args[0])
	}
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
		return usageError(err.Error())
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
		return usageError(err.Error())
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
