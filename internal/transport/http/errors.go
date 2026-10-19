package http

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("not found")

// Error codes returned by the HTTP API.
const (
	CodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeAuthTokenMissing       = "AUTH_TOKEN_MISSING"
	CodeAuthTokenInvalid       = "AUTH_TOKEN_INVALID"
	CodeAuthTokenExpired       = "AUTH_TOKEN_EXPIRED"
	CodeAuthEmailTaken         = "AUTH_EMAIL_TAKEN"
	CodeValidationFailed       = "VALIDATION_FAILED"
	CodeBadRequest             = "BAD_REQUEST"
	CodeNotFound               = "NOT_FOUND"
	CodeGuestNotFound          = "GUEST_NOT_FOUND"
	CodeMethodNotAllowed       = "METHOD_NOT_ALLOWED"
	CodeInternal               = "INTERNAL_ERROR"
)

// Field is a single validation error field.
type Field struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrResponse is the standard error response body.
type ErrResponse struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Fields  []Field `json:"fields,omitempty"`
}

// TokenResponse is the login response body.
type TokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
}

// EnvResponse is the registration response body.
type EnvResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// GuestResponse is a single guest response body.
type GuestResponse struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	RoomNumber string `json:"room_number"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Note       string `json:"note"`
	UserID     int64  `json:"user_id"`
}

// GuestListResponse is a paginated guest list response body.
type GuestListResponse struct {
	Items []GuestResponse `json:"items"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Total int             `json:"total"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, code string, status int, message string, fields ...Field) {
	writeJSON(w, status, ErrResponse{Code: code, Message: message, Fields: fields})
}
