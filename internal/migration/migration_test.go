package migration

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

func TestManagerApplyStatusAndDown(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	files := fstest.MapFS{
		"demo/20260615000000_init.up.sql":      {Data: []byte(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);`)},
		"demo/20260615000000_init.down.sql":    {Data: []byte(`DROP TABLE items;`)},
		"demo/20260615000001_seed.up.sql":      {Data: []byte(`INSERT INTO items (id, name) VALUES (1, 'first');`)},
		"demo/20260615000001_seed.down.sql":    {Data: []byte(`DELETE FROM items WHERE id = 1;`)},
		"demo/20260615000002_add_col.up.sql":   {Data: []byte(`ALTER TABLE items ADD COLUMN description TEXT NOT NULL DEFAULT '';`)},
		"demo/20260615000002_add_col.down.sql": {Data: []byte(`CREATE TABLE items_next (id INTEGER PRIMARY KEY, name TEXT NOT NULL); INSERT INTO items_next (id, name) SELECT id, name FROM items; DROP TABLE items; ALTER TABLE items_next RENAME TO items;`)},
	}
	manager := NewManager(db, files, "demo")
	ctx := context.Background()

	pending, err := manager.Pending(ctx)
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(pending) != 3 {
		t.Fatalf("pending length = %d, want 3", len(pending))
	}
	ran, err := manager.Apply(ctx)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(ran) != 3 {
		t.Fatalf("ran length = %d, want 3", len(ran))
	}
	if err := manager.CheckNoPending(ctx); err != nil {
		t.Fatalf("check pending: %v", err)
	}
	statuses, err := manager.Status(ctx)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	for _, status := range statuses {
		if !status.Applied {
			t.Fatalf("status %s applied = false", status.Version)
		}
	}
	rolledBack, err := manager.Down(ctx)
	if err != nil {
		t.Fatalf("down: %v", err)
	}
	if rolledBack == nil || rolledBack.Version != "20260615000002" {
		t.Fatalf("rolled back = %+v, want 20260615000002", rolledBack)
	}
	pending, err = manager.Pending(ctx)
	if err != nil {
		t.Fatalf("pending after down: %v", err)
	}
	if len(pending) != 1 || pending[0].Version != "20260615000002" {
		t.Fatalf("pending after down = %+v", pending)
	}
}
