package http

import (
	"errors"
	"net/http"
	"strconv"

	"guest-accounting/internal/domain"
	"guest-accounting/internal/repository/postgres"
	"guest-accounting/internal/rop"
)

const defaultPage = 1
const defaultLimit = 20

func (h *Handler) listGuests(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(w, r)
	if !ok {
		return
	}
	page, limit := parsePagination(r)
	repo := postgres.NewGuestRepository(h.db)
	list, err := repo.List(userID, page, limit)
	if err != nil {
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toGuestListResponse(list))
}

func (h *Handler) createGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(w, r)
	if !ok {
		return
	}
	g, err := readBody[domain.Guest](r)
	if err != nil {
		writeError(w, CodeBadRequest, http.StatusBadRequest, "invalid request body")
		return
	}
	g.ID = 0
	g.UserID = userID

	validated := domain.ValidateGuest(g)
	if de, ok := leftDomainError(validated); ok {
		writeDomainValidation(w, de)
		return
	}

	repo := postgres.NewGuestRepository(h.db)
	id, err := repo.Create(g)
	if err != nil {
		writeInternalError(w)
		return
	}
	g.ID = id
	writeJSON(w, http.StatusCreated, toGuestResponse(g))
}

func (h *Handler) searchGuests(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(w, r)
	if !ok {
		return
	}
	query := r.URL.Query().Get("q")
	page, limit := parsePagination(r)
	repo := postgres.NewGuestRepository(h.db)
	list, err := repo.Search(strconv.FormatInt(userID, 10), query, page, limit)
	if err != nil {
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, toGuestListResponse(list))
}

func (h *Handler) getGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(w, r)
	if !ok {
		return
	}
	id, ok := guestIDOr404(w, r)
	if !ok {
		return
	}
	repo := postgres.NewGuestRepository(h.db)
	g, ok := getGuestOr404(w, repo, id, userID)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, toGuestResponse(g))
}

func (h *Handler) updateGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(w, r)
	if !ok {
		return
	}
	id, ok := guestIDOr404(w, r)
	if !ok {
		return
	}
	repo := postgres.NewGuestRepository(h.db)
	existing, ok := getGuestOr404(w, repo, id, userID)
	if !ok {
		return
	}
	patch, err := readBody[domain.Guest](r)
	if err != nil {
		writeError(w, CodeBadRequest, http.StatusBadRequest, "invalid request body")
		return
	}
	if patch.FullName != "" {
		existing.FullName = patch.FullName
	}
	if patch.Phone != "" {
		existing.Phone = patch.Phone
	}
	if patch.RoomNumber != "" {
		existing.RoomNumber = patch.RoomNumber
	}
	if patch.CheckIn != "" {
		existing.CheckIn = patch.CheckIn
	}
	if patch.CheckOut != "" {
		existing.CheckOut = patch.CheckOut
	}
	if patch.Note != "" {
		existing.Note = patch.Note
	}
	existing.ID = id
	existing.UserID = userID

	validated := domain.ValidateGuest(existing)
	if de, ok := leftDomainError(validated); ok {
		writeDomainValidation(w, de)
		return
	}
	updated, ok := updateOr404(w, repo, id, userID, existing)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, toGuestResponse(updated))
}

func (h *Handler) deleteGuest(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromCtx(w, r)
	if !ok {
		return
	}
	id, ok := guestIDOr404(w, r)
	if !ok {
		return
	}
	repo := postgres.NewGuestRepository(h.db)
	if !deleteOr404(w, repo, id, userID) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// userIDFromCtx returns the user id from the request context, writing a 401
// response when the token is missing.
func userIDFromCtx(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, CodeAuthTokenMissing, http.StatusUnauthorized, "token missing")
		return 0, false
	}
	return userID, true
}

// guestIDOr404 parses the guest id from the path, writing a 404 on failure.
func guestIDOr404(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := guestIDFromPath(r)
	if err != nil {
		writeError(w, CodeGuestNotFound, http.StatusNotFound, "guest not found")
		return 0, false
	}
	return id, true
}

// writeInternalError writes a generic 500 JSON error without leaking details.
func writeInternalError(w http.ResponseWriter) {
	writeError(w, CodeInternal, http.StatusInternalServerError, "internal error")
}

// getGuestOr404 fetches a guest by id/user, writing 404 on not found or 500 on
// other errors. Returns false when a response was already written.
func getGuestOr404(w http.ResponseWriter, repo postgres.GuestRepository, id, userID int64) (domain.Guest, bool) {
	g, err := repo.Get(id, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			writeError(w, CodeGuestNotFound, http.StatusNotFound, "guest not found")
			return domain.Guest{}, false
		}
		writeInternalError(w)
		return domain.Guest{}, false
	}
	return g, true
}

// writeDomainValidation writes a 422 validation error from a DomainError.
func writeDomainValidation(w http.ResponseWriter, de domain.DomainError) {
	fields := make([]Field, 0, len(de.Fields))
	for _, f := range de.Fields {
		fields = append(fields, Field{Field: f.Field, Message: f.Message})
	}
	writeError(w, CodeValidationFailed, http.StatusUnprocessableEntity, de.Message, fields...)
}

// updateOr404 updates a guest, writing 404 on not found or 500 on other errors.
func updateOr404(w http.ResponseWriter, repo postgres.GuestRepository, id, userID int64, g domain.Guest) (domain.Guest, bool) {
	updated, err := repo.Update(id, userID, g)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			writeError(w, CodeGuestNotFound, http.StatusNotFound, "guest not found")
			return domain.Guest{}, false
		}
		writeInternalError(w)
		return domain.Guest{}, false
	}
	return updated, true
}

// deleteOr404 deletes a guest, writing 404 on not found or 500 on other errors.
func deleteOr404(w http.ResponseWriter, repo postgres.GuestRepository, id, userID int64) bool {
	if err := repo.Delete(id, userID); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			writeError(w, CodeGuestNotFound, http.StatusNotFound, "guest not found")
			return false
		}
		writeInternalError(w)
		return false
	}
	return true
}

func leftDomainError(e rop.Either[domain.DomainError, domain.Guest]) (domain.DomainError, bool) {
	var de domain.DomainError
	got := false
	rop.Fold(e,
		func(err domain.DomainError) struct{} { de = err; got = true; return struct{}{} },
		func(domain.Guest) struct{} { return struct{}{} },
	)
	return de, got
}

func parsePagination(r *http.Request) (page, limit int) {
	page, limit = defaultPage, defaultLimit
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && limit > 0 {
			page = n/limit + 1
		}
	}
	return
}

func toGuestListResponse(list domain.GuestList) GuestListResponse {
	items := make([]GuestResponse, 0, len(list.Items))
	for _, g := range list.Items {
		items = append(items, toGuestResponse(g))
	}
	return GuestListResponse{Items: items, Page: list.Page, Limit: list.Limit, Total: list.Total}
}

func guestIDFromPath(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, ErrNotFound
	}
	return strconv.ParseInt(idStr, 10, 64)
}

func toGuestResponse(g domain.Guest) GuestResponse {
	return GuestResponse{
		ID:         g.ID,
		FullName:   g.FullName,
		Phone:      g.Phone,
		RoomNumber: g.RoomNumber,
		CheckIn:    g.CheckIn,
		CheckOut:   g.CheckOut,
		Note:       g.Note,
		UserID:     g.UserID,
	}
}
