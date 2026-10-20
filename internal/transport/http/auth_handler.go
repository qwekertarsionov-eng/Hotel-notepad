package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"guest-accounting/internal/domain"
)

func (h *Handler) authRegister(w http.ResponseWriter, r *http.Request) {
	body, err := readBody[map[string]string](r)
	if err != nil {
		writeError(w, CodeBadRequest, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := domain.ValidateEmail(body["email"]); err != nil {
		writeValidationError(w, "email", err.Error())
		return
	}
	if err := domain.ValidatePassword(body["password"]); err != nil {
		writeValidationError(w, "password", err.Error())
		return
	}
	id, err := h.auth.Register(body["email"], body["password"])
	if err != nil {
		respondDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, EnvResponse{ID: id, Email: body["email"]})
}

func (h *Handler) authLogin(w http.ResponseWriter, r *http.Request) {
	body, err := readBody[map[string]string](r)
	if err != nil {
		writeError(w, CodeBadRequest, http.StatusBadRequest, "invalid request body")
		return
	}
	token, err := h.auth.Login(body["email"], body["password"])
	if err != nil {
		respondDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{Token: token, TokenType: "Bearer", ExpiresIn: 3600})
}

func readBody[T any](r *http.Request) (T, error) {
	var v T
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return v, err
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}

func writeValidationError(w http.ResponseWriter, field, message string) {
	writeError(w, CodeValidationFailed, http.StatusUnprocessableEntity, "validation failed", Field{Field: field, Message: message})
}

// respondDomainError maps a service DomainError to an HTTP status, code and
// safe client message and writes the JSON error response. Unknown errors become
// 500 without leaking details. This is the single place where DomainError is
// translated into an HTTP status.
func respondDomainError(w http.ResponseWriter, err error) {
	code, status, message := mapDomainError(err)
	writeError(w, code, status, message)
}

// domainErrorHandler maps a DomainError to an HTTP status, code and message.
type domainErrorHandler func(de domain.DomainError) (code string, status int, message string)

// domainErrorHandlers is a lookup table keyed by DomainError code. Unknown
// codes fall through to the default 500 internal error below.
var domainErrorHandlers = map[string]domainErrorHandler{
	CodeAuthEmailTaken: func(de domain.DomainError) (string, int, string) {
		return CodeAuthEmailTaken, http.StatusConflict, de.Message
	},
	CodeAuthInvalidCredentials: func(de domain.DomainError) (string, int, string) {
		return CodeAuthInvalidCredentials, http.StatusUnauthorized, de.Message
	},
}

// mapDomainError converts a service DomainError into an HTTP status, code and
// safe client message. Unknown errors become 500 without leaking details.
func mapDomainError(err error) (code string, status int, message string) {
	var de domain.DomainError
	if errors.As(err, &de) {
		if handler, ok := domainErrorHandlers[de.Code]; ok {
			return handler(de)
		}
	}
	return CodeInternal, http.StatusInternalServerError, "internal error"
}
