package postgres

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func insertUser(t *testing.T, db *sql.DB, email string) int64 {
	var id int64
	if err := db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, "hash").Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertGuestRaw(t *testing.T, db *sql.DB, query string, args ...interface{}) error {
	_, err := db.Exec(query, args...)
	return err
}

func TestTST_DB_003(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)
	insertUser(t, db, "a@example.com")
	err := insertGuestRaw(t, db, "INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6, $7)", 999999, "Гость", "+7", "1", "2026-05-01", "2026-05-05", "")
	if err == nil {
		t.Fatal("guest with nonexistent user_id must be rejected by the foreign key")
	}
	if pgCode(err) != "23503" {
		t.Fatalf("expected foreign_key_violation (23503), got %s: %v", pgCode(err), err)
	}
}

func TestTST_DB_004(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)
	userID := insertUser(t, db, "a@example.com")
	err := insertGuestRaw(t, db, "INSERT INTO guests (user_id, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6)", userID, "+7", "1", "2026-05-01", "2026-05-05", "")
	if err == nil {
		t.Fatal("NOT NULL violation must be rejected")
	}
	if pgCode(err) != "23502" {
		t.Fatalf("expected not_null_violation (23502), got %s: %v", pgCode(err), err)
	}
}

func TestTST_DB_005(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)
	userID := insertUser(t, db, "a@example.com")
	err := insertGuestRaw(t, db, "INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6, $7)", userID, "Гость", "+7", "1", "2026-05-05", "2026-05-01", "")
	if err == nil {
		t.Fatal("check_out <= check_in must be rejected by the CHECK constraint")
	}
	if pgCode(err) != "23514" {
		t.Fatalf("expected check_violation (23514), got %s: %v", pgCode(err), err)
	}
}

func TestTST_DB_006(t *testing.T) {
	db := openTestDB(t)
	applyUpMigrations(t, db)
	userA := insertUser(t, db, "a@example.com")
	userB := insertUser(t, db, "b@example.com")
	for i := 0; i < 50; i++ {
		if _, err := db.Exec("INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6, $7)", userA, "Гость А", "+7", "1", "2026-05-01", "2026-05-05", ""); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 5000; i++ {
		if _, err := db.Exec("INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note) VALUES ($1, $2, $3, $4, $5, $6, $7)", userB, "Гость Б", "+7", "1", "2026-05-01", "2026-05-05", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("ANALYZE guests"); err != nil {
		t.Fatal(err)
	}
	var plan strings.Builder
	rows, err := db.Query("EXPLAIN (ANALYZE) SELECT * FROM guests WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20", userA)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatal(err)
		}
		plan.WriteString(line)
		plan.WriteString("\n")
	}
	text := plan.String()
	if !strings.Contains(text, "ix_guests_user_created_at") {
		t.Fatalf("EXPLAIN ANALYZE must use ix_guests_user_created_at index, got plan:\n%s", text)
	}
	if strings.Contains(text, "Seq Scan") {
		t.Fatalf("EXPLAIN ANALYZE must not use Seq Scan, got plan:\n%s", text)
	}
}
