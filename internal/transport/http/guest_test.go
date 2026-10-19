package http

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

type testGuest struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	RoomNumber string `json:"room_number"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Note       string `json:"note"`
	UserID     int64  `json:"user_id"`
}

type testGuestList struct {
	Items []testGuest `json:"items"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int         `json:"total"`
}

func createGuest(t *testing.T, srv *httptest.Server, token string, body map[string]interface{}) *httptest.ResponseRecorder {
	return execReq(t, srv, "POST", "/api/v1/guests", body, token)
}

func parseGuest(t *testing.T, rec *httptest.ResponseRecorder) testGuest {
	var g testGuest
	if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil {
		t.Fatalf("response is not JSON: %s", rec.Body.String())
	}
	return g
}

func parseList(t *testing.T, rec *httptest.ResponseRecorder) testGuestList {
	var l testGuestList
	if err := json.Unmarshal(rec.Body.Bytes(), &l); err != nil {
		t.Fatalf("response is not JSON: %s", rec.Body.String())
	}
	return l
}

func guestBody(fullName string) map[string]interface{} {
	return map[string]interface{}{
		"full_name":   fullName,
		"phone":       "+7 999 123-45-67",
		"room_number": "101",
		"check_in":    "2026-05-01",
		"check_out":   "2026-05-05",
		"note":        "early check-in",
		"user_id":     9999,
	}
}

func TestTST_GUEST_001(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, uid := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := createGuest(t, srv, token, guestBody("Иванов Иван Иванович"))
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	g := parseGuest(t, rec)
	if g.ID == 0 || g.UserID != uid || g.FullName != "Иванов Иван Иванович" {
		t.Fatalf("unexpected guest: %+v", g)
	}
	var owner int64
	if err := db.QueryRow("SELECT user_id FROM guests WHERE id = $1", g.ID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner != uid {
		t.Fatal("user_id must equal the token owner, not the body value")
	}
}

func TestTST_GUEST_002(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	body := guestBody("")
	rec := createGuest(t, srv, token, body)
	if rec.Code != 422 {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "VALIDATION_FAILED" || e.Message == "" {
		t.Fatalf("expected VALIDATION_FAILED, got %+v", e)
	}
	found := false
	for _, f := range e.Fields {
		if f.Field == "full_name" {
			found = true
		}
	}
	if !found {
		t.Fatal("validation fields must contain full_name")
	}
}

func TestTST_GUEST_003(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	body := guestBody("Иванов Иван Иванович")
	body["check_in"] = "2026-05-05"
	body["check_out"] = "2026-05-01"
	rec := createGuest(t, srv, token, body)
	if rec.Code != 422 {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "VALIDATION_FAILED" {
		t.Fatalf("expected VALIDATION_FAILED, got %+v", e)
	}
}

func TestTST_GUEST_004(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	rec := createGuest(t, srv, "", guestBody("Иванов Иван Иванович"))
	if rec.Code != 401 {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTST_GUEST_005(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	tokenA, uidA := signup(t, srv, "a@example.com", "secret-pass-1")
	tokenB, _ := signup(t, srv, "b@example.com", "secret-pass-1")
	createGuest(t, srv, tokenA, guestBody("Гость А один"))
	createGuest(t, srv, tokenA, guestBody("Гость А два"))
	createGuest(t, srv, tokenB, guestBody("Гость Б"))

	rec := execReq(t, srv, "GET", "/api/v1/guests", nil, tokenA)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	l := parseList(t, rec)
	if l.Page != 1 || l.Limit != 20 || l.Total != 2 || len(l.Items) != 2 {
		t.Fatalf("unexpected pagination: %+v", l)
	}
	for _, item := range l.Items {
		if item.UserID != uidA {
			t.Fatal("list must contain only the owner records")
		}
	}
}

func TestTST_GUEST_006(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	createGuest(t, srv, token, guestBody("Иванов Иван Иванович"))
	createGuest(t, srv, token, guestBody("Петров Петр Петрович"))

	rec := execReq(t, srv, "GET", "/api/v1/guests/search?q=Иванов", nil, token)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	l := parseList(t, rec)
	if len(l.Items) != 1 || l.Items[0].FullName != "Иванов Иван Иванович" {
		t.Fatalf("search must find the record by substring: %+v", l)
	}
}

func TestTST_GUEST_007(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := execReq(t, srv, "GET", "/api/v1/guests/search?q=НеизвестноеИмя", nil, token)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	l := parseList(t, rec)
	if len(l.Items) != 0 {
		t.Fatal("search without results must return empty items")
	}
}

func TestTST_GUEST_008(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := createGuest(t, srv, token, guestBody("Иванов Иван Иванович"))
	g := parseGuest(t, rec)

	rec = execReq(t, srv, "GET", "/api/v1/guests/"+fmtID(g.ID), nil, token)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	g2 := parseGuest(t, rec)
	if g2.ID != g.ID {
		t.Fatal("view must return the requested record")
	}
}

func TestTST_GUEST_009(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	tokenA, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	tokenB, _ := signup(t, srv, "b@example.com", "secret-pass-1")
	rec := createGuest(t, srv, tokenB, guestBody("Гость Б"))
	gB := parseGuest(t, rec)

	rec = execReq(t, srv, "GET", "/api/v1/guests/"+fmtID(gB.ID), nil, tokenA)
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "GUEST_NOT_FOUND" {
		t.Fatalf("expected GUEST_NOT_FOUND, got %+v", e)
	}
}

func TestTST_GUEST_010(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := createGuest(t, srv, token, guestBody("Иванов Иван Иванович"))
	g := parseGuest(t, rec)

	update := map[string]interface{}{"note": "updated note", "phone": "+7 000 000-00-00"}
	rec = execReq(t, srv, "PUT", "/api/v1/guests/"+fmtID(g.ID), update, token)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	updated := parseGuest(t, rec)
	if updated.Note != "updated note" || updated.Phone != "+7 000 000-00-00" {
		t.Fatal("passed fields must be updated")
	}
	if updated.FullName != g.FullName || updated.RoomNumber != g.RoomNumber {
		t.Fatal("non-passed fields must remain unchanged")
	}
}

func TestTST_GUEST_011(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	tokenA, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	tokenB, _ := signup(t, srv, "b@example.com", "secret-pass-1")
	rec := createGuest(t, srv, tokenB, guestBody("Гость Б"))
	gB := parseGuest(t, rec)

	rec = execReq(t, srv, "PUT", "/api/v1/guests/"+fmtID(gB.ID), map[string]interface{}{"note": "hack"}, tokenA)
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = execReq(t, srv, "GET", "/api/v1/guests/"+fmtID(gB.ID), nil, tokenB)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	unchanged := parseGuest(t, rec)
	if unchanged.Note == "hack" {
		t.Fatal("someone else's record must not be modified")
	}
}

func TestTST_GUEST_012(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := createGuest(t, srv, token, guestBody("Иванов Иван Иванович"))
	g := parseGuest(t, rec)

	rec = execReq(t, srv, "DELETE", "/api/v1/guests/"+fmtID(g.ID), nil, token)
	if rec.Code != 204 {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatal("204 must have no body")
	}
	rec = execReq(t, srv, "GET", "/api/v1/guests/"+fmtID(g.ID), nil, token)
	if rec.Code != 404 {
		t.Fatalf("repeat GET must return 404, got %d", rec.Code)
	}
}

func TestTST_GUEST_013(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	tokenA, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	tokenB, _ := signup(t, srv, "b@example.com", "secret-pass-1")
	rec := createGuest(t, srv, tokenB, guestBody("Гость Б"))
	gB := parseGuest(t, rec)

	rec = execReq(t, srv, "DELETE", "/api/v1/guests/"+fmtID(gB.ID), nil, tokenA)
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = execReq(t, srv, "GET", "/api/v1/guests/"+fmtID(gB.ID), nil, tokenB)
	if rec.Code != 200 {
		t.Fatalf("someone else's record must still exist, got %d", rec.Code)
	}
}

func TestTST_GUEST_014(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	token, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	rec := execReq(t, srv, "GET", "/api/v1/guests/0", nil, token)
	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	e := parseErr(t, rec)
	if e.Code != "GUEST_NOT_FOUND" {
		t.Fatalf("expected GUEST_NOT_FOUND, got %+v", e)
	}
}

func TestTST_GUEST_015(t *testing.T) {
	db := openTestDB(t)
	applyMigrations(t, db)
	t.Cleanup(func() { truncateTables(t, db) })
	srv := newServer(t, db)

	tokenA, _ := signup(t, srv, "a@example.com", "secret-pass-1")
	tokenB, _ := signup(t, srv, "b@example.com", "secret-pass-1")
	createGuest(t, srv, tokenA, guestBody("Гость А"))
	recB := createGuest(t, srv, tokenB, guestBody("Гость Б"))
	gB := parseGuest(t, recB)

	rec := execReq(t, srv, "GET", "/api/v1/guests", nil, tokenA)
	l := parseList(t, rec)
	for _, item := range l.Items {
		if item.ID == gB.ID {
			t.Fatal("user A's list must not contain user B's record")
		}
	}
	rec = execReq(t, srv, "GET", "/api/v1/guests/search?q=Гость", nil, tokenA)
	l = parseList(t, rec)
	for _, item := range l.Items {
		if item.ID == gB.ID {
			t.Fatal("user A's search must not contain user B's record")
		}
	}
}

func fmtID(id int64) string {
	return fmt.Sprintf("%d", id)
}
