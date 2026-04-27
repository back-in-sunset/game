package middleware

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"platform/api/internal/types"

	"github.com/golang-jwt/jwt/v4"
)

type JwtAuthMiddleware struct {
	publicKey *rsa.PublicKey
}

func NewJwtAuthMiddleware(publicKey *rsa.PublicKey) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{publicKey: publicKey}
}

func (m *JwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimSpace(r.Header.Get("Authorization"))
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("invalid signing method")
			}
			return m.publicKey, nil
		})
		if err != nil || !parsedToken.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		uid, ok := claims["uid"]
		if !ok {
			http.Error(w, "missing uid in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), types.UserIDKey, uid)
		r = r.WithContext(ctx)
		next(w, r)
	}
}
