package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type ctxKey int

const userIDKey ctxKey = 1

// UserIDFromContext returns the user id stored in ctx by the JWT middleware.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}

// JWTMiddleware validates the Bearer JWT on protected routes.
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("JWT_SECRET")
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, CodeAuthTokenMissing, http.StatusUnauthorized, "token missing")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			writeError(w, CodeAuthTokenInvalid, http.StatusUnauthorized, "invalid token")
			return
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(parts[0] + "." + parts[1]))
		expected := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(expected), []byte(parts[2])) {
			writeError(w, CodeAuthTokenInvalid, http.StatusUnauthorized, "invalid token")
			return
		}
		payload, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
		if err != nil {
			writeError(w, CodeAuthTokenInvalid, http.StatusUnauthorized, "invalid token")
			return
		}
		var claims struct {
			Sub string `json:"sub"`
			Exp int64  `json:"exp"`
		}
		if err := json.Unmarshal(payload, &claims); err != nil {
			writeError(w, CodeAuthTokenInvalid, http.StatusUnauthorized, "invalid token")
			return
		}
		if claims.Exp != 0 && time.Now().Unix() >= claims.Exp {
			writeError(w, CodeAuthTokenExpired, http.StatusUnauthorized, "token expired")
			return
		}
		var userID int64
		if _, err := fmt.Sscanf(claims.Sub, "%d", &userID); err != nil {
			writeError(w, CodeAuthTokenInvalid, http.StatusUnauthorized, "invalid token subject")
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
		next.ServeHTTP(w, r)
	})
}
