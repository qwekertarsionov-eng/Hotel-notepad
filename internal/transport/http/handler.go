package http

import (
	"database/sql"
	"net/http"
	"time"

	"guest-accounting/internal/config"
	"guest-accounting/internal/repository/postgres"
	"guest-accounting/internal/service"
)

// Handler routes HTTP requests to auth and guest handlers.
type Handler struct {
	db     *sql.DB
	auth   service.AuthService
	guests service.GuestService
}

// NewHandler creates an HTTP handler backed by db.
func NewHandler(db *sql.DB) (*Handler, error) {
	cfg, _ := config.Load()
	return &Handler{
		db:     db,
		auth:   service.NewAuthService(postgres.NewUserRepository(db), cfg.JWT_SECRET, 24*time.Hour),
		guests: service.NewGuestService(),
	}, nil
}

// Routes returns the HTTP handler for all API routes.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", h.register)
	mux.HandleFunc("/api/v1/auth/login", h.login)
	mux.Handle("/api/v1/guests", JWTMiddleware(http.HandlerFunc(h.guestsRoute)))
	mux.Handle("/api/v1/guests/search", JWTMiddleware(http.HandlerFunc(h.searchRoute)))
	mux.Handle("/api/v1/guests/{id}", JWTMiddleware(http.HandlerFunc(h.guestByIDRoute)))
	mux.HandleFunc("/", h.notFound)
	return mux
}

func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, CodeNotFound, http.StatusNotFound, "not found")
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, CodeMethodNotAllowed, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	h.authRegister(w, r)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	h.authLogin(w, r)
}

func (h *Handler) guestsRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listGuests(w, r)
	case http.MethodPost:
		h.createGuest(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) searchRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	h.searchGuests(w, r)
}

func (h *Handler) guestByIDRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getGuest(w, r)
	case http.MethodPut:
		h.updateGuest(w, r)
	case http.MethodDelete:
		h.deleteGuest(w, r)
	default:
		methodNotAllowed(w)
	}
}
