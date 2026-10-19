package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"guest-accounting/internal/domain"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// GuestRepository defines the data access operations for guests.
// Every query filters by user_id to guarantee user data isolation (REQ-GUEST-007).
type GuestRepository interface {
	Create(g domain.Guest) (int64, error)
	Get(id, userID int64) (domain.Guest, error)
	List(userID int64, page, limit int) (domain.GuestList, error)
	Search(userID, query string, page, limit int) (domain.GuestList, error)
	Update(id, userID int64, g domain.Guest) (domain.Guest, error)
	Delete(id, userID int64) error
}

// UserRepository defines the data access operations for users.
type UserRepository interface {
	Create(email, passwordHash string) (int64, error)
	GetByEmail(email string) (domain.User, error)
	GetByID(id int64) (domain.User, error)
}

type guestRepository struct {
	db *sql.DB
}

// NewGuestRepository creates a guest repository backed by db.
func NewGuestRepository(db *sql.DB) GuestRepository {
	return &guestRepository{db: db}
}

const guestColumns = "id, user_id, full_name, phone, room_number, check_in, check_out, note, created_at"

const defaultLimit = 20
const maxLimit = 100

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

// scannable is implemented by *sql.Row and *sql.Rows.
type scannable interface {
	Scan(dest ...any) error
}

func scanGuest(row scannable) (domain.Guest, error) {
	var g domain.Guest
	var checkIn, checkOut time.Time
	err := row.Scan(
		&g.ID, &g.UserID, &g.FullName, &g.Phone, &g.RoomNumber,
		&checkIn, &checkOut, &g.Note, &g.CreatedAt,
	)
	if err != nil {
		return domain.Guest{}, err
	}
	g.CheckIn = checkIn.Format("2006-01-02")
	g.CheckOut = checkOut.Format("2006-01-02")
	return g, nil
}

func (r *guestRepository) Create(g domain.Guest) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO guests (user_id, full_name, phone, room_number, check_in, check_out, note)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		g.UserID, g.FullName, g.Phone, g.RoomNumber, g.CheckIn, g.CheckOut, g.Note,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *guestRepository) Get(id, userID int64) (domain.Guest, error) {
	g, err := scanGuest(r.db.QueryRow(
		`SELECT `+guestColumns+` FROM guests WHERE id = $1 AND user_id = $2`,
		id, userID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Guest{}, ErrNotFound
		}
		return domain.Guest{}, err
	}
	return g, nil
}

func (r *guestRepository) List(userID int64, page, limit int) (domain.GuestList, error) {
	page, limit = normalizePagination(page, limit)
	offset := (page - 1) * limit

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM guests WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return domain.GuestList{}, err
	}

	rows, err := r.db.Query(
		`SELECT `+guestColumns+` FROM guests WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return domain.GuestList{}, err
	}
	defer rows.Close()

	items := make([]domain.Guest, 0, limit)
	for rows.Next() {
		g, err := scanGuest(rows)
		if err != nil {
			return domain.GuestList{}, err
		}
		items = append(items, g)
	}
	if err := rows.Err(); err != nil {
		return domain.GuestList{}, err
	}
	return domain.GuestList{Items: items, Page: page, Limit: limit, Total: total}, nil
}

func (r *guestRepository) Search(userID, query string, page, limit int) (domain.GuestList, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return domain.GuestList{}, fmt.Errorf("invalid user_id: %w", err)
	}
	page, limit = normalizePagination(page, limit)
	offset := (page - 1) * limit
	pattern := "%" + query + "%"

	where := `WHERE user_id = $1 AND (full_name ILIKE $2 OR phone ILIKE $2 OR room_number ILIKE $2)`

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM guests `+where, uid, pattern).Scan(&total); err != nil {
		return domain.GuestList{}, err
	}

	rows, err := r.db.Query(
		`SELECT `+guestColumns+` FROM guests `+where+` ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		uid, pattern, limit, offset,
	)
	if err != nil {
		return domain.GuestList{}, err
	}
	defer rows.Close()

	items := make([]domain.Guest, 0, limit)
	for rows.Next() {
		g, err := scanGuest(rows)
		if err != nil {
			return domain.GuestList{}, err
		}
		items = append(items, g)
	}
	if err := rows.Err(); err != nil {
		return domain.GuestList{}, err
	}
	return domain.GuestList{Items: items, Page: page, Limit: limit, Total: total}, nil
}

func (r *guestRepository) Update(id, userID int64, g domain.Guest) (domain.Guest, error) {
	updated, err := scanGuest(r.db.QueryRow(
		`UPDATE guests
		 SET full_name = $2, phone = $3, room_number = $4, check_in = $5, check_out = $6, note = $7, updated_at = now()
		 WHERE id = $1 AND user_id = $8
		 RETURNING `+guestColumns,
		id, g.FullName, g.Phone, g.RoomNumber, g.CheckIn, g.CheckOut, g.Note, userID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Guest{}, ErrNotFound
		}
		return domain.Guest{}, err
	}
	return updated, nil
}

func (r *guestRepository) Delete(id, userID int64) error {
	res, err := r.db.Exec(`DELETE FROM guests WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a user repository backed by db.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(email, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userRepository) GetByEmail(email string) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *userRepository) GetByID(id int64) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}
