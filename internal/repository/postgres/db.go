package postgres

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect opens a connection pool to PostgreSQL using the pgx stdlib driver
// and verifies connectivity with Ping.
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
