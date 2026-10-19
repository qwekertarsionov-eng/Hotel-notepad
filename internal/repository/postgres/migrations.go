package postgres

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const migrationDir = "../../../migrations"

// Up applies all versioned .up.sql migrations in ascending order, each in a transaction.
func Up(db *sql.DB) error {
	return applyMigrations(db, ".up.sql", false)
}

// Down reverts all versioned .down.sql migrations in descending order, each in a transaction.
func Down(db *sql.DB) error {
	return applyMigrations(db, ".down.sql", true)
}

func applyMigrations(db *sql.DB, suffix string, reverse bool) error {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	if reverse {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}
	for _, f := range files {
		if err := applyFile(db, filepath.Join(migrationDir, f)); err != nil {
			return err
		}
	}
	return nil
}

func applyFile(db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, stmt := range splitSQL(string(data)) {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
