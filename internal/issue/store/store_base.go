package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/db/migrations"
	"github.com/version-1/tasq/internal/migration"
)

type Store struct {
	db                *sql.DB
	attachmentStorage *AttachmentStorage
}

const (
	IssueSortByID        IssueSortBy = "id"
	IssueSortByPriority  IssueSortBy = "priority"
	IssueSortByCreatedAt IssueSortBy = "created_at"
	IssueSortByUpdatedAt IssueSortBy = "updated_at"
)

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
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
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	if _, err := migration.NewManager(db, migrations.Files, "issue-tracker").Apply(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) SetAttachmentStorage(storage *AttachmentStorage) {
	s.attachmentStorage = storage
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) checkMigrations(ctx context.Context) error {
	if err := migration.NewManager(s.db, migrations.Files, "issue-tracker").CheckNoPending(ctx); err != nil {
		return fmt.Errorf("issue tracker sqlite migrations pending: %w", err)
	}
	return nil
}
