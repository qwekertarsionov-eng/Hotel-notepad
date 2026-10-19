package postgres

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestTST_DB_007(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec("INSERT INTO users (email, password_hash) VALUES ($1, $2)", "tx@example.com", "hash")
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	_, err = tx.Exec("INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6, $7)", 999999, "Гость", "+7", "1", "2026-05-01", "2026-05-05", "")
	if err == nil {
		tx.Rollback()
		t.Fatal("mid-transaction FK violation must cause rollback")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM guests").Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	if cnt != 0 {
		t.Fatalf("rollback must leave no partial records, got %d guests", cnt)
	}

	var userID int64
	if err := db.QueryRow("SELECT id FROM users WHERE email = $1", "tx@example.com").Scan(&userID); err != nil {
		t.Fatal(err)
	}
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec("INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6, $7)", userID, "Гость", "+7", "1", "2026-05-01", "2026-05-05", "")
	if err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM guests WHERE user_id = $1", userID).Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Fatalf("commit must persist the record, got %d guests", cnt)
	}
}
