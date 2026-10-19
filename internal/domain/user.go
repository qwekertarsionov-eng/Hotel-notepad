package domain

// User is a registered user.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
}
