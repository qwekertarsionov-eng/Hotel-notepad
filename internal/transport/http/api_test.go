package http

import (
	"bytes"
	"encoding/json"
	"nethttp"
	"nethttp/httptest"
	"os"
	"strings"
	"testing"
)

type testStatusResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestTST_API_001(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": "x@example.com", "password": "secret-pass-1"}, "")
	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := createGuest(t, srv, token, guestBody("Иванов Иван Иванович"))
	g := parseGuest(t, rec)

	cases := []struct {
		method string
		path   string
		body   interface{}
		token  string
		want   int
	}{
		{"POST", "/api/v1/auth/register", map[string]string{"email": "x@example.com", "password": "secret-pass-1"}, "", 201},
		{"POST", "/api/v1/auth/login", map[string]string{"email": "a@example.com", "password": "secret-pass-1"}, "", 200},
		{"GET", "/api/v1/guests", nil, token, 200},
		{"POST", "/api/v1/guests", guestBody("Гость 2"), token, 201},
		{"GET", "/api/v1/guests/search?q=Иванов", nil, token, 200},
		{"GET", "/api/v1/guests/" + fmtID(g.ID), nil, token, 200},
		{"PUT", "/api/v1/guests/" + fmtID(g.ID), map[string]interface{}{"note": "n"}, token, 200},
		{"DELETE", "/api/v1/guests/" + fmtID(g.ID), nil, token, 204},
	}
	for _, c := range cases {
		rec := execReq(t, srv, c.method, c.path, c.body, c.token)
		if rec.Code != c.want {
			t.Fatalf("%s %s: expected %d, got %d: %s", c.method, c.path, c.want, rec.Code, rec.Body.String())
		}
	}
}

func TestTST_API_002(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	signup(t, srv, "a@example.com", "secret-pass-1")
	rec := execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": "a@example.com", "password": "secret-pass-1"}, "")
	if rec.Code != 409 {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	e := parseErr(t, rec)
	if e.Code == "" || e.Message == "" {
		t.Fatal("every error must contain code and message")
	}
	rec = execReq(t, srv, "GET", "/api/v1/guests", nil, "")
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	e = parseErr(t, rec)
	if e.Code == "" || e.Message == "" {
		t.Fatal("every error must contain code and message")
	}
}

func TestTST_API_003(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	req, _ := nethttp.NewRequest(nethttp.MethodPost, srv.URL+"/api/v1/auth/register", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %+v", e)
	}
}

func TestTST_API_004(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	rec := execReq(t, srv, "GET", "/api/v1/no-such-path", nil, "")
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %+v", e)
	}
}

func TestTST_API_005(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	rec := execReq(t, srv, "GET", "/api/v1/auth/register", nil, "")
	if rec.Code != 405 {
		t.Fatalf("expected 405, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("expected METHOD_NOT_ALLOWED, got %+v", e)
	}
}

func TestTST_SEC_003(t *testing.T) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		t.Skip("JWT_SECRET is not set")
	}
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	rec := execReq(t, srv, "GET", "/api/v1/no-such-path", nil, "")
	body := rec.Body.String()
	if strings.Contains(body, secret) {
		t.Fatal("response must not contain JWT_SECRET")
	}
	rec = execReq(t, srv, "POST", "/api/v1/auth/register", map[string]string{"email": "x", "password": "y"}, "")
	if strings.Contains(rec.Body.String(), secret) {
		t.Fatal("response must not contain JWT_SECRET")
	}
}

func TestTST_SEC_004(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	srv := newServer(t, db)
	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	rec := execReq(t, srv, "POST", "/api/v1/guests", guestBody("Иванов Иван Иванович"), token)
	if rec.Code != 500 {
		t.Fatalf("expected 500 when DB is unavailable, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	var e testStatusResp
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("500 response must be JSON: %s", body)
	}
	if e.Code == "" || e.Message == "" {
		t.Fatal("500 response must contain code and message")
	}
	forbidden := []string{"sql", "SQL", "panic", "stack", "trace", "users", "guests", "migration", "pgx", "postgres", "/internal/", "SELECT", "INSERT", "DELETE"}
	lower := strings.ToLower(body)
	for _, f := range forbidden {
		if strings.Contains(lower, strings.ToLower(f)) {
			t.Fatalf("500 body must not contain internal detail %q: %s", f, body)
		}
	}
}
