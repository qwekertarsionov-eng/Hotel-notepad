package http

import (
	"errors"
	"net/http"
	"testing"

	"guest-accounting/internal/domain"
)

func TestMapDomainError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantCode   string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "AUTH_EMAIL_TAKEN maps to 409",
			err:        domain.DomainError{Code: CodeAuthEmailTaken, Message: "taken"},
			wantCode:   CodeAuthEmailTaken,
			wantStatus: http.StatusConflict,
			wantMsg:    "taken",
		},
		{
			name:       "AUTH_INVALID_CREDENTIALS maps to 401",
			err:        domain.DomainError{Code: CodeAuthInvalidCredentials, Message: "bad creds"},
			wantCode:   CodeAuthInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "bad creds",
		},
		{
			name:       "unknown code falls back to 500",
			err:        domain.DomainError{Code: "UNKNOWN_CODE", Message: "something"},
			wantCode:   CodeInternal,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal error",
		},
		{
			name:       "non-DomainError falls back to 500",
			err:        errors.New("some error"),
			wantCode:   CodeInternal,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal error",
		},
		{
			name:       "nil error falls back to 500",
			err:        nil,
			wantCode:   CodeInternal,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, status, msg := mapDomainError(tt.err)
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if msg != tt.wantMsg {
				t.Errorf("message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}
