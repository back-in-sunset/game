package jwtx

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

// GetToken 生成 token
func GetToken(privateKey *rsa.PrivateKey, userID int64, expire int64) (string, error) {
	now := time.Now().Unix()

	claims := make(jwt.MapClaims)
	claims["exp"] = now + expire
	claims["iat"] = now
	claims["uid"] = userID

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(privateKey)
}

// ParseToken 解析 token
func ParseToken(tokenString string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	parser := jwt.Parser{UseJSONNumber: true}
	return parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 检查 alg
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return publicKey, nil
	})
}
