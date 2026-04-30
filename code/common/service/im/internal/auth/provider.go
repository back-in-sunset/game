package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"

	"im/internal/domain"

	"github.com/golang-jwt/jwt/v4"
)

type Principal struct {
	UserID int64
	Domain domain.IMDomain
	Scope  domain.Scope
}

type AuthProvider interface {
	Authenticate(token string, imDomain domain.IMDomain, scope domain.Scope) (Principal, error)
}

type JWTProvider struct {
	publicKey *rsa.PublicKey
}

func NewJWTProvider(publicKeyFile string) (*JWTProvider, error) {
	data, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(data)
	if err != nil {
		return nil, err
	}
	return &JWTProvider{publicKey: key}, nil
}

func (p *JWTProvider) Authenticate(token string, imDomain domain.IMDomain, scope domain.Scope) (Principal, error) {
	token = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(token), "Bearer "))
	if token == "" {
		return Principal{}, errors.New("missing token")
	}

	parsed, err := jwt.Parse(token, func(parsed *jwt.Token) (interface{}, error) {
		if _, ok := parsed.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.publicKey, nil
	})
	if err != nil || !parsed.Valid {
		return Principal{}, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Principal{}, errors.New("invalid claims")
	}

	uidValue, ok := claims["uid"]
	if !ok {
		return Principal{}, errors.New("missing uid")
	}

	uidFloat, ok := uidValue.(float64)
	if !ok {
		return Principal{}, errors.New("uid has unsupported type")
	}

	if err := scope.Validate(imDomain); err != nil {
		return Principal{}, err
	}

	return Principal{
		UserID: int64(uidFloat),
		Domain: imDomain,
		Scope:  scope.Normalize(),
	}, nil
}
