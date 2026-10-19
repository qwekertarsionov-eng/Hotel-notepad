package postgres

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func migrationFiles() ([]string, bool) {
	entries, err := os.ReadDir("../../../migrations")
	if err != nil {
		return nil, false
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	if len(files) == 0 {
		return nil, false
	}
	sort.Strings(files)
	return files, true
}

func applyUpMigrations(t *testing.T, db *sql.DB) {
	files, ok := migrationFiles()
	if !ok {
		t.Skip("no migration files found")
	}
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join("../../../migrations", f))
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		for _, stmt := range splitSQL(string(data)) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("migration %s: %v", f, err)
			}
		}
	}
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1", name).Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	return cnt > 0
}

func TestTST_DB_001(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)
	if !tableExists(t, db, "users") || !tableExists(t, db, "guests") {
		t.Fatal("migrations must create users and guests tables")
	}
}

func TestTST_DB_002(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)
	applyUpMigrations(t, db)
	if !tableExists(t, db, "users") || !tableExists(t, db, "guests") {
		t.Fatal("re-applied migrations must not break the schema")
	}
}
