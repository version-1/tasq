package tq

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/db/migrations"
	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/migration"
)

type migrateTarget struct {
	Name string
	Path string
	Dir  string
}

type migrateResult struct {
	Database   string          `json:"database"`
	Path       string          `json:"path"`
	Applied    []string        `json:"applied,omitempty"`
	RolledBack string          `json:"rolled_back,omitempty"`
	Statuses   []migrateStatus `json:"statuses,omitempty"`
}

type migrateStatus struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
}

func (a app) routeMigrate(ctx context.Context, args []string, cfg config) error {
	if len(args) > 1 {
		return usageError("usage: tq migrate [down|status]")
	}
	action := "up"
	if len(args) == 1 {
		action = args[0]
	}
	switch action {
	case "up":
		return a.migrateUp(ctx, cfg)
	case "down":
		return a.migrateDown(ctx, cfg)
	case "status":
		return a.migrateStatus(ctx, cfg)
	case "help", "-help", "--help":
		printMigrateHelp(a.stdout)
		return nil
	default:
		return usageError("unknown migrate action %q", action)
	}
}

func (a app) migrateUp(ctx context.Context, cfg config) error {
	results, err := withMigrationTargets(ctx, func(ctx context.Context, target migrateTarget, manager migration.Manager) (migrateResult, error) {
		applied, err := manager.Apply(ctx)
		if err != nil {
			return migrateResult{}, err
		}
		result := migrateResult{Database: target.Name, Path: target.Path}
		for _, item := range applied {
			result.Applied = append(result.Applied, migrationLabel(item.Version, item.Name))
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	return writeMigrateResults(a.stdout, cfg.output, "Applied migrations", results)
}

func (a app) migrateDown(ctx context.Context, cfg config) error {
	results, err := withMigrationTargets(ctx, func(ctx context.Context, target migrateTarget, manager migration.Manager) (migrateResult, error) {
		rolledBack, err := manager.Down(ctx)
		if err != nil {
			return migrateResult{}, err
		}
		result := migrateResult{Database: target.Name, Path: target.Path}
		if rolledBack != nil {
			result.RolledBack = migrationLabel(rolledBack.Version, rolledBack.Name)
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	return writeMigrateResults(a.stdout, cfg.output, "Rolled back migrations", results)
}

func (a app) migrateStatus(ctx context.Context, cfg config) error {
	results, err := withMigrationTargets(ctx, func(ctx context.Context, target migrateTarget, manager migration.Manager) (migrateResult, error) {
		statuses, err := manager.Status(ctx)
		if err != nil {
			return migrateResult{}, err
		}
		result := migrateResult{Database: target.Name, Path: target.Path}
		for _, status := range statuses {
			result.Statuses = append(result.Statuses, migrateStatus{
				Version: status.Version,
				Name:    status.Name,
				Applied: status.Applied,
			})
		}
		return result, nil
	})
	if err != nil {
		return err
	}
	return writeMigrateResults(a.stdout, cfg.output, "Migration status", results)
}

func withMigrationTargets(ctx context.Context, fn func(context.Context, migrateTarget, migration.Manager) (migrateResult, error)) ([]migrateResult, error) {
	home, err := tqconfig.EnsureHome()
	if err != nil {
		return nil, err
	}
	targets := []migrateTarget{
		{Name: "issue-tracker", Path: tqconfig.IssueTrackerDBPath(home), Dir: "issue-tracker"},
		{Name: "orchestrator", Path: tqconfig.OrchestratorDBPath(home), Dir: "orchestrator"},
	}
	results := make([]migrateResult, 0, len(targets))
	for _, target := range targets {
		if err := os.MkdirAll(filepath.Dir(target.Path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
		db, err := sql.Open("sqlite", target.Path)
		if err != nil {
			return nil, fmt.Errorf("open %s sqlite: %w", target.Name, err)
		}
		db.SetMaxOpenConns(1)
		result, runErr := fn(ctx, target, migration.NewManager(db, migrations.Files, target.Dir))
		closeErr := db.Close()
		if runErr != nil {
			return nil, fmt.Errorf("%s migrations: %w", target.Name, runErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s sqlite: %w", target.Name, closeErr)
		}
		results = append(results, result)
	}
	return results, nil
}

func checkMigrationTargetsNoPending(ctx context.Context) error {
	var pending []string
	_, err := withMigrationTargets(ctx, func(ctx context.Context, target migrateTarget, manager migration.Manager) (migrateResult, error) {
		migrations, err := manager.Pending(ctx)
		if err != nil {
			return migrateResult{}, err
		}
		for _, item := range migrations {
			pending = append(pending, target.Name+":"+migrationLabel(item.Version, item.Name))
		}
		return migrateResult{Database: target.Name, Path: target.Path}, nil
	})
	if err != nil {
		return err
	}
	if len(pending) > 0 {
		return fmt.Errorf("pending migrations: %s; run `tq migrate` before starting services", strings.Join(pending, ", "))
	}
	return nil
}

func checkOrchestratorMigrationTargetNoPending(ctx context.Context, home string) error {
	target := migrateTarget{Name: "orchestrator", Path: tqconfig.OrchestratorDBPath(home), Dir: "orchestrator"}
	if err := os.MkdirAll(filepath.Dir(target.Path), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", target.Path)
	if err != nil {
		return fmt.Errorf("open orchestrator sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	pending, err := migration.NewManager(db, migrations.Files, target.Dir).Pending(ctx)
	if err != nil {
		return fmt.Errorf("orchestrator migrations: %w", err)
	}
	if len(pending) == 0 {
		return nil
	}
	labels := make([]string, 0, len(pending))
	for _, item := range pending {
		labels = append(labels, target.Name+":"+migrationLabel(item.Version, item.Name))
	}
	return fmt.Errorf("pending migrations: %s; run `tq migrate` before starting services", strings.Join(labels, ", "))
}

func migrationLabel(version string, name string) string {
	return version + "_" + name
}

func printMigrateHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq migrate [down|status]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  tq migrate         Apply all pending migrations for local databases")
	fmt.Fprintln(w, "  tq migrate down    Roll back one migration per local database")
	fmt.Fprintln(w, "  tq migrate status  List applied and pending migrations")
}
