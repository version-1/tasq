package runstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/db/migrations"
	"github.com/version-1/tasq/internal/migration"
)

type Store struct {
	db *sql.DB
}

var ErrProjectHasRunningRuns = errors.New("project has running runs")

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.checkMigrations(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func OpenMigrated(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := migration.NewManager(db, migrations.Files, "orchestrator").Apply(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) checkMigrations(ctx context.Context) error {
	if err := migration.NewManager(s.db, migrations.Files, "orchestrator").CheckNoPending(ctx); err != nil {
		return fmt.Errorf("orchestrator sqlite migrations pending: %w", err)
	}
	return nil
}
