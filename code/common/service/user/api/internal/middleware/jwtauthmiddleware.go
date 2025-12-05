package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"user/api/internal/types"
	"user/utils/jwtx"

	"github.com/golang-jwt/jwt"
)

type JwtAuthMiddleware struct {
	publicKey *rsa.PublicKey
}

func NewJwtAuthMiddleware(publicKey *rsa.PublicKey) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{
		publicKey: publicKey,
	}
}

func (m *JwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		parsedToken, err := jwtx.ParseToken(tokenStr, m.publicKey)
		if err != nil || !parsedToken.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			types.UserIDKey,
			parsedToken.Claims.(jwt.MapClaims)["uid"].(json.Number),
		)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
