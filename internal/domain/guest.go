package domain

import "time"

// Guest is a guest record owned by a user.
type Guest struct {
	ID         int64
	UserID     int64
	FullName   string
	Phone      string
	RoomNumber string
	CheckIn    string
	CheckOut   string
	Note       string
	CreatedAt  time.Time
}

// GuestList is a paginated list of guests.
type GuestList struct {
	Items []Guest
	Page  int
	Limit int
	Total int
}
