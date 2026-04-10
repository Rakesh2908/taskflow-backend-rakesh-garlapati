package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ctxUserIDKey contextKey = "user_id"
	ctxEmailKey  contextKey = "email"
)

type authClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if authz == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authz, prefix) || len(authz) <= len(prefix) {
				writeJSONError(w, http.StatusUnauthorized, "malformed authorization header")
				return
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(authz, prefix))
			if tokenStr == "" {
				writeJSONError(w, http.StatusUnauthorized, "malformed authorization header")
				return
			}

			claims := &authClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenSignatureInvalid
				}
				return secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil || token == nil || !token.Valid {
				writeJSONError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			if claims.UserID == "" || claims.Email == "" {
				writeJSONError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			uid, err := uuid.Parse(claims.UserID)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserIDKey, uid)
			ctx = context.WithValue(ctx, ctxEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserIDKey)
	uid, ok := v.(uuid.UUID)
	return uid, ok
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

