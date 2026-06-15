package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

const schemaMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

type Status struct {
	Version string
	Name    string
	Applied bool
}

type Manager struct {
	db         *sql.DB
	files      fs.FS
	migrations string
}

func NewManager(db *sql.DB, files fs.FS, migrations string) Manager {
	return Manager{db: db, files: files, migrations: migrations}
}

func (m Manager) Apply(ctx context.Context) ([]Migration, error) {
	migrations, err := m.load()
	if err != nil {
		return nil, err
	}
	if err := m.ensureSchemaMigrations(ctx); err != nil {
		return nil, err
	}
	applied, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	ran := []Migration{}
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}
		if err := m.applyOne(ctx, migration); err != nil {
			return nil, err
		}
		ran = append(ran, migration)
	}
	return ran, nil
}

func (m Manager) Down(ctx context.Context) (*Migration, error) {
	migrations, err := m.load()
	if err != nil {
		return nil, err
	}
	if err := m.ensureSchemaMigrations(ctx); err != nil {
		return nil, err
	}
	applied, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	byVersion := map[string]Migration{}
	for _, migration := range migrations {
		byVersion[migration.Version] = migration
	}
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if !applied[migration.Version] {
			continue
		}
		if _, ok := byVersion[migration.Version]; !ok {
			return nil, fmt.Errorf("applied migration %q has no migration file", migration.Version)
		}
		if err := m.downOne(ctx, migration); err != nil {
			return nil, err
		}
		return &migration, nil
	}
	return nil, nil
}

func (m Manager) Status(ctx context.Context) ([]Status, error) {
	migrations, err := m.load()
	if err != nil {
		return nil, err
	}
	applied, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	statuses := make([]Status, 0, len(migrations))
	for _, migration := range migrations {
		statuses = append(statuses, Status{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: applied[migration.Version],
		})
	}
	return statuses, nil
}

func (m Manager) Pending(ctx context.Context) ([]Migration, error) {
	migrations, err := m.load()
	if err != nil {
		return nil, err
	}
	applied, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	pending := []Migration{}
	for _, migration := range migrations {
		if !applied[migration.Version] {
			pending = append(pending, migration)
		}
	}
	return pending, nil
}

func (m Manager) CheckNoPending(ctx context.Context) error {
	pending, err := m.Pending(ctx)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	versions := make([]string, 0, len(pending))
	for _, migration := range pending {
		versions = append(versions, migration.Version+"_"+migration.Name)
	}
	return fmt.Errorf("pending migrations: %s; run `tq migrate` before starting services", strings.Join(versions, ", "))
}

func (m Manager) load() ([]Migration, error) {
	entries, err := fs.ReadDir(m.files, m.migrations)
	if err != nil {
		return nil, fmt.Errorf("read migrations %s: %w", m.migrations, err)
	}
	type pair struct {
		up   string
		down string
	}
	pairs := map[string]pair{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			key := strings.TrimSuffix(name, ".up.sql")
			item := pairs[key]
			item.up = name
			pairs[key] = item
		case strings.HasSuffix(name, ".down.sql"):
			key := strings.TrimSuffix(name, ".down.sql")
			item := pairs[key]
			item.down = name
			pairs[key] = item
		}
	}
	keys := make([]string, 0, len(pairs))
	for key, pair := range pairs {
		if pair.up == "" || pair.down == "" {
			return nil, fmt.Errorf("migration %s must have both up and down files", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	migrations := make([]Migration, 0, len(keys))
	for _, key := range keys {
		version, name, ok := strings.Cut(key, "_")
		if !ok || version == "" || name == "" {
			return nil, fmt.Errorf("migration %s must be named <version>_<name>", key)
		}
		pair := pairs[key]
		upSQL, err := fs.ReadFile(m.files, path.Join(m.migrations, pair.up))
		if err != nil {
			return nil, err
		}
		downSQL, err := fs.ReadFile(m.files, path.Join(m.migrations, pair.down))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		})
	}
	return migrations, nil
}

func (m Manager) ensureSchemaMigrations(ctx context.Context) error {
	if _, err := m.db.ExecContext(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func (m Manager) applied(ctx context.Context) (map[string]bool, error) {
	if ok, err := m.schemaMigrationsExists(ctx); err != nil {
		return nil, err
	} else if !ok {
		return map[string]bool{}, nil
	}
	rows, err := m.db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()
	applied := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return applied, nil
}

func (m Manager) schemaMigrationsExists(ctx context.Context) (bool, error) {
	var count int
	err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("inspect schema_migrations: %w", err)
	}
	return count > 0, nil
}

func (m Manager) applyOne(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("apply migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, name) VALUES (?, ?)`, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("record migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (m Manager) downOne(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, migration.DownSQL); err != nil {
		return fmt.Errorf("rollback migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = ?`, migration.Version); err != nil {
		return fmt.Errorf("remove migration %s_%s: %w", migration.Version, migration.Name, err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func rollback(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return
	}
}
