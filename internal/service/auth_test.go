package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"guest-accounting/internal/domain"
	"guest-accounting/internal/repository/postgres"
)

type fakeUserRepository struct {
	users map[string]domain.User
	next  int64
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: make(map[string]domain.User), next: 1}
}

func (f *fakeUserRepository) Create(email, passwordHash string) (int64, error) {
	if _, exists := f.users[email]; exists {
		return 0, &duplicateEmailError{email: email}
	}
	id := f.next
	f.next++
	f.users[email] = domain.User{ID: id, Email: email, PasswordHash: passwordHash}
	return id, nil
}

func (f *fakeUserRepository) GetByEmail(email string) (domain.User, error) {
	u, ok := f.users[email]
	if !ok {
		return domain.User{}, postgres.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepository) GetByID(id int64) (domain.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, postgres.ErrNotFound
}

type duplicateEmailError struct{ email string }

func (e *duplicateEmailError) Error() string {
	return `pq: duplicate key value violates unique constraint "ux_users_email"`
}

func TestAuthRegisterSuccess(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewAuthService(repo, "test-secret", time.Hour)

	id, err := svc.Register("a@example.com", "secret-pass-1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	u, err := repo.GetByEmail("a@example.com")
	if err != nil || u.ID != id {
		t.Fatalf("unexpected stored user: %+v, %v", u, err)
	}
	if u.PasswordHash == "secret-pass-1" {
		t.Fatal("plaintext password must not be stored")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("secret-pass-1")); err != nil {
		t.Fatalf("stored hash does not match password: %v", err)
	}
}

func TestAuthRegisterSamePasswordDifferentHash(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewAuthService(repo, "test-secret", time.Hour)

	if _, err := svc.Register("a@example.com", "secret-pass-1"); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	h1 := repo.users["a@example.com"].PasswordHash
	if _, err := svc.Register("b@example.com", "secret-pass-1"); err != nil {
		t.Fatalf("second Register: %v", err)
	}
	h2 := repo.users["b@example.com"].PasswordHash
	if h1 == h2 {
		t.Fatal("identical passwords must produce different hashes (salt)")
	}
}

func TestAuthRegisterEmailTaken(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewAuthService(repo, "test-secret", time.Hour)

	if _, err := svc.Register("a@example.com", "secret-pass-1"); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	_, err := svc.Register("a@example.com", "secret-pass-1")
	var de domain.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T: %v", err, err)
	}
	if de.Code != "AUTH_EMAIL_TAKEN" || de.Message == "" {
		t.Fatalf("expected AUTH_EMAIL_TAKEN, got %+v", de)
	}
}

func TestAuthLoginSuccess(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewAuthService(repo, "test-secret", 24*time.Hour)

	id, err := svc.Register("a@example.com", "secret-pass-1")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	token, err := svc.Login("a@example.com", "secret-pass-1")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := verifyJWT(token, []byte("test-secret"))
	if err != nil {
		t.Fatalf("verifyJWT: %v", err)
	}
	if claims["sub"] != strconv.FormatInt(id, 10) {
		t.Fatalf("unexpected sub: %v", claims["sub"])
	}
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) <= time.Now().Unix() {
		t.Fatalf("token exp is not in the future: %v", exp)
	}
}

func TestAuthLoginWrongPassword(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewAuthService(repo, "test-secret", time.Hour)
	if _, err := svc.Register("a@example.com", "secret-pass-1"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	_, err := svc.Login("a@example.com", "wrong-password")
	var de domain.DomainError
	if !errors.As(err, &de) || de.Code != "AUTH_INVALID_CREDENTIALS" || de.Message == "" {
		t.Fatalf("expected AUTH_INVALID_CREDENTIALS, got %+v", de)
	}
}

func TestAuthLoginUnknownEmail(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewAuthService(repo, "test-secret", time.Hour)

	_, err := svc.Login("nonexistent@example.com", "secret-pass-1")
	var de domain.DomainError
	if !errors.As(err, &de) || de.Code != "AUTH_INVALID_CREDENTIALS" || de.Message == "" {
		t.Fatalf("expected AUTH_INVALID_CREDENTIALS, got %+v", de)
	}
}

func TestAuthHashNotPlaintext(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("p@ssw0rd"), bcryptCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	if string(hash) == "p@ssw0rd" {
		t.Fatal("hash must not equal plaintext")
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("p@ssw0rd")); err != nil {
		t.Fatalf("CompareHashAndPassword: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrong")); err == nil {
		t.Fatal("wrong password must not match hash")
	}
}

func TestAuthJWTInvalidSignature(t *testing.T) {
	token, err := signJWT([]byte("secret-a"), map[string]any{"sub": "1", "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatalf("signJWT: %v", err)
	}
	if _, err := verifyJWT(token, []byte("secret-b")); err == nil {
		t.Fatal("token signed with different secret must not verify")
	}
}

func TestAuthJWTExpired(t *testing.T) {
	token, err := signJWT([]byte("secret"), map[string]any{"sub": "1", "exp": time.Now().Add(-time.Hour).Unix()})
	if err != nil {
		t.Fatalf("signJWT: %v", err)
	}
	claims, err := verifyJWT(token, []byte("secret"))
	if err != nil {
		t.Fatalf("verifyJWT: %v", err)
	}
	exp, _ := claims["exp"].(float64)
	if int64(exp) >= time.Now().Unix() {
		t.Fatal("expected expired token")
	}
}

func b64uDecode(s string) ([]byte, error) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
}

func verifyTokenParts(token string, secret []byte) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errors.New("invalid signature")
	}
	payload, err := b64uDecode(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
