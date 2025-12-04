package svc

import (
	"crypto/rsa"
	"os"
	"user/api/internal/config"
	"user/api/internal/middleware"
	"user/api/userclient"

	"github.com/golang-jwt/jwt"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	UserRpc userclient.User

	JwtAuth    rest.Middleware
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 加载非对称密钥
	privateKeyData, err := os.ReadFile(c.Auth.PrivateKeyFile)
	if err != nil {
		panic(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		panic(err)
	}
	publicKeyData, err := os.ReadFile(c.Auth.PublicKeyFile)
	if err != nil {
		panic(err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:     c,
		UserRpc:    userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		JwtAuth:    middleware.NewJwtAuthMiddleware(publicKey).Handle,
	}
}

type Validater interface {
	SendCode(receiver string, code string) error
	Validate(code string) (bool, error)
}

type EmailValidater struct {
	Email string
}

type SmsValidater struct {
	Mobile string
}
