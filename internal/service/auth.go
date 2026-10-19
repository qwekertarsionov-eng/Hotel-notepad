package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"guest-accounting/internal/domain"
	"guest-accounting/internal/repository/postgres"
)

// AuthService handles registration and login.
type AuthService interface {
	Register(email, password string) (int64, error)
	Login(email, password string) (string, error)
}

type authService struct {
	users     postgres.UserRepository
	jwtSecret []byte
	expiry    time.Duration
}

// NewAuthService creates an auth service backed by the user repository.
// Tokens are HS256 JWTs signed with jwtSecret and valid for expiry.
func NewAuthService(users postgres.UserRepository, jwtSecret string, expiry time.Duration) AuthService {
	return &authService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		expiry:    expiry,
	}
}

const bcryptCost = bcrypt.DefaultCost

// Register creates a user with email and bcrypt-hashed password.
// Returns the new user id or a DomainError (AUTH_EMAIL_TAKEN / INTERNAL_ERROR).
func (s *authService) Register(email, password string) (int64, error) {
	if _, err := s.users.GetByEmail(email); err == nil {
		return 0, domain.DomainError{
			Code:    "AUTH_EMAIL_TAKEN",
			Message: "user with this email already exists",
		}
	} else if !errors.Is(err, postgres.ErrNotFound) {
		return 0, domain.DomainError{
			Code:    "INTERNAL_ERROR",
			Message: "failed to check email uniqueness",
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return 0, domain.DomainError{
			Code:    "INTERNAL_ERROR",
			Message: "failed to hash password",
		}
	}

	id, err := s.users.Create(email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return 0, domain.DomainError{
				Code:    "AUTH_EMAIL_TAKEN",
				Message: "user with this email already exists",
			}
		}
		return 0, domain.DomainError{
			Code:    "INTERNAL_ERROR",
			Message: "failed to create user",
		}
	}
	return id, nil
}

// Login verifies email/password and issues a JWT.
// Returns the token or a DomainError (AUTH_INVALID_CREDENTIALS / INTERNAL_ERROR).
func (s *authService) Login(email, password string) (string, error) {
	user, err := s.users.GetByEmail(email)
	if err != nil {
		// Nonexistent email and DB errors produce the same client-facing error
		// so the response does not reveal whether the email exists.
		return "", domain.DomainError{
			Code:    "AUTH_INVALID_CREDENTIALS",
			Message: "invalid email or password",
		}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.DomainError{
			Code:    "AUTH_INVALID_CREDENTIALS",
			Message: "invalid email or password",
		}
	}
	return s.issueToken(user.ID)
}

func (s *authService) issueToken(userID int64) (string, error) {
	now := time.Now()
	claims := map[string]any{
		"sub": strconv.FormatInt(userID, 10),
		"exp": now.Add(s.expiry).Unix(),
		"iat": now.Unix(),
	}
	return signJWT(s.jwtSecret, claims)
}

func b64u(b []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// signJWT signs claims as an HS256 JWT (header {"alg":"HS256","typ":"JWT"}).
func signJWT(secret []byte, claims map[string]any) (string, error) {
	header := b64u([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := b64u(payloadBytes)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(header + "." + payload))
	return header + "." + payload + "." + b64u(mac.Sum(nil)), nil
}

// verifyJWT validates the HS256 signature and returns the claims.
func verifyJWT(token string, secret []byte) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := b64u(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errors.New("invalid signature")
	}
	payload, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
