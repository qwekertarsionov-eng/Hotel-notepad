package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"nethttp"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type testEnv struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type testTokenResp struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
}

type testErrResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fields  []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"fields"`
}

func openTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func applyMigrations(t *testing.T, db *sql.DB) {
	entries, err := os.ReadDir("../../../migrations")
	if err != nil {
		t.Skip("migrations directory is not readable")
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	if len(files) == 0 {
		t.Skip("no migration files found")
	}
	sort.Strings(files)
	for _, f := range files {
		data, err := os.ReadFile("migrations/" + f)
		if err != nil {
			data, err = os.ReadFile("../../../migrations/" + f)
			if err != nil {
				t.Fatalf("read migration %s: %v", f, err)
			}
		}
		for _, stmt := range splitSQL(string(data)) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("migration %s: %v", f, err)
			}
		}
	}
}

func truncateTables(t *testing.T, db *sql.DB) {
	if _, err := db.Exec("DELETE FROM guests"); err != nil {
		t.Fatalf("truncate guests: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("truncate users: %v", err)
	}
}

func newServer(t *testing.T, db *sql.DB) *httptest.Server {
	h, err := NewHandler(db)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return httptest.NewServer(h.Routes())
}

func execReq(t *testing.T, srv *httptest.Server, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var r io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(data)
	}
	req, _ := nethttp.NewRequest(method, srv.URL+path, r)
	if r != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	srv.Config.Handler.ServeHTTP(rec, req)
	return rec
}

func parseErr(t *testing.T, rec *httptest.ResponseRecorder) testErrResp {
	var e testErrResp
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("response is not JSON: %s", rec.Body.String())
	}
	return e
}

func signup(t *testing.T, srv *httptest.Server, email, password string) (string, int64) {
	rec := execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": email, "password": password}, "")
	if rec.Code != 201 {
		t.Fatalf("signup: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var e testEnv
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	rec = execReq(t, srv, "POST", "/api/v1/auth/login", map[string]string{"email": email, "password": password}, "")
	if rec.Code != 200 {
		t.Fatalf("login: expected 200, got %d", rec.Code)
	}
	var tr testTokenResp
	if err := json.Unmarshal(rec.Body.Bytes(), &tr); err != nil {
		t.Fatal(err)
	}
	return tr.Token, e.ID
}

func b64u(b []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

func makeToken(secret, sub string, exp int64) string {
	h := b64u([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p := b64u([]byte(fmt.Sprintf(`{"sub":"%s","exp":%d}`, sub, exp)))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(h + "." + p))
	return h + "." + p + "." + b64u(mac.Sum(nil))
}

func jwtSecret(t *testing.T) string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		t.Skip("JWT_SECRET is not set")
	}
	return secret
}

func TestTST_AUTH_001(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	rec := execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": "a@example.com", "password": "secret-pass-1"}, "")
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var e testEnv
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	if e.ID == 0 || e.Email != "a@example.com" {
		t.Fatalf("unexpected registration body: %s", rec.Body.String())
	}
	var cnt int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 users row, got %d", cnt)
	}
}

func TestTST_AUTH_002(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	body := map[string]string{"email": "a@example.com", "password": "secret-pass-1"}
	execReq(t, srv, "POST", "/api/v1/auth/register", body, "")
	rec := execReq(t, srv, "POST", "/api/v1/auth/register", body, "")
	if rec.Code != 409 {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "AUTH_EMAIL_TAKEN" || e.Message == "" {
		t.Fatalf("expected AUTH_EMAIL_TAKEN with message, got %+v", e)
	}
}

func TestTST_AUTH_003(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	for _, body := range []map[string]string{
		{"email": "not-email", "password": "secret-pass-1"},
		{"email": "a@example.com", "password": "short"},
		{"email": "", "password": "secret-pass-1"},
		{"email": "a@example.com", "password": ""},
	} {
		rec := execReq(t, srv, "POST", "/api/v1/auth/register", body, "")
		if rec.Code != 422 {
			t.Fatalf("expected 422 for %v, got %d: %s", body, rec.Code, rec.Body.String())
		}
		e := parseErr(t, rec)
		if e.Code != "VALIDATION_FAILED" || e.Message == "" || len(e.Fields) == 0 {
			t.Fatalf("expected VALIDATION_FAILED with fields, got %+v", e)
		}
	}
}

func TestTST_AUTH_004(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	if token == "" {
		t.Fatal("expected a token in the login response")
	}
	rec := execReq(t, srv, "POST", "/api/v1/auth/login", map[string]string{"email": "a@example.com", "password": "secret-pass-1"}, "")
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var tr testTokenResp
	if err := json.Unmarshal(rec.Body.Bytes(), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.Token == "" || tr.TokenType != "Bearer" || tr.ExpiresIn <= 0 {
		t.Fatalf("unexpected token response: %+v", tr)
	}
}

func TestTST_AUTH_005(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	signup(t, srv, "a@example.com", "secret-pass-1")
	rec := execReq(t, srv, "POST", "/api/v1/auth/login", map[string]string{"email": "a@example.com", "password": "wrong-password"}, "")
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "AUTH_INVALID_CREDENTIALS" {
		t.Fatalf("expected AUTH_INVALID_CREDENTIALS, got %+v", e)
	}
}

func TestTST_AUTH_006(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	signup(t, srv, "a@example.com", "secret-pass-1")
	rec := execReq(t, srv, "POST", "/api/v1/auth/login", map[string]string{"email": "nonexistent@example.com", "password": "secret-pass-1"}, "")
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "AUTH_INVALID_CREDENTIALS" || e.Message == "" {
		t.Fatalf("expected AUTH_INVALID_CREDENTIALS with message, got %+v", e)
	}
}

func TestTST_AUTH_007(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	rec := execReq(t, srv, "GET", "/api/v1/guests", nil, "")
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "AUTH_TOKEN_MISSING" {
		t.Fatalf("expected AUTH_TOKEN_MISSING, got %+v", e)
	}
}

func TestTST_AUTH_008(t *testing.T) {
	jwtSecret(t)
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token := makeToken("wrong-secret", "1", time.Now().Add(time.Hour).Unix())
	rec := execReq(t, srv, "GET", "/api/v1/guests", nil, token)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "AUTH_TOKEN_INVALID" {
		t.Fatalf("expected AUTH_TOKEN_INVALID, got %+v", e)
	}
}

func TestTST_AUTH_009(t *testing.T) {
	jwtSecret(t)
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token := makeToken(jwtSecret(t), "1", time.Now().Add(-time.Hour).Unix())
	rec := execReq(t, srv, "GET", "/api/v1/guests", nil, token)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "AUTH_TOKEN_EXPIRED" {
		t.Fatalf("expected AUTH_TOKEN_EXPIRED, got %+v", e)
	}
}

func TestTST_AUTH_010(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	password := "secret-pass-1"
	rec := execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": "a@example.com", "password": password}, "")
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["password"]; ok {
		t.Fatal("plaintext password must not be in the response")
	}
	if _, ok := body["password_hash"]; ok {
		t.Fatal("password_hash must not be in the response")
	}
	var hash string
	if err := db.QueryRow("SELECT password_hash FROM users WHERE email = $1", "a@example.com").Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if hash == "" || hash == password {
		t.Fatal("password must be stored as a hash, not as plaintext")
	}
}

func TestTST_SEC_001(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	password := "secret-pass-1"
	execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": "a@example.com", "password": password}, "")
	var hash string
	if err := db.QueryRow("SELECT password_hash FROM users WHERE email = $1", "a@example.com").Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(hash, password) {
		t.Fatal("plaintext password must not be stored in the database")
	}
}

func TestTST_SEC_002(t *testing.T) {
	jwtSecret(t)
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	signup(t, srv, "a@example.com", "secret-pass-1")
	token := makeToken("wrong-secret", "1", time.Now().Add(time.Hour).Unix())
	rec := execReq(t, srv, "GET", "/api/v1/guests", nil, token)
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
