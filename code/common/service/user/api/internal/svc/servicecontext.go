package svc

import (
	"crypto/rsa"
	"os"
	"time"
	"user/api/internal/config"
	"user/api/internal/middleware"
	"user/api/internal/verify"
	"user/api/userclient"
	"user/model"

	"github.com/golang-jwt/jwt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	UserRpc          userclient.User
	UserModel        model.UserModel
	UserProfileModel model.UserProfileModel
	EmailVerifier    verify.CodeVerifier
	SmsVerifier      verify.CodeVerifier
	CodeRateLimiter  verify.RateLimiter

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

	limiter := verify.NewMemoryRateLimiter()

	return &ServiceContext{
		Config:           c,
		UserRpc:          userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		UserModel:        model.NewUserModel(sqlx.NewMysql(c.Mysql.DataSource), c.CacheRedis),
		UserProfileModel: model.NewUserProfileModel(sqlx.NewMysql(c.Mysql.DataSource), c.CacheRedis),
		CodeRateLimiter:  limiter,
		EmailVerifier: verify.NewEmailVerifier(verify.EmailConfig{
			Host:       c.Verify.Email.Host,
			Port:       c.Verify.Email.Port,
			Username:   c.Verify.Email.Username,
			Password:   c.Verify.Email.Password,
			From:       c.Verify.Email.From,
			Subject:    c.Verify.Email.Subject,
			TTL:        time.Duration(c.Verify.CodeTTLSeconds) * time.Second,
			Cooldown:   time.Duration(c.Verify.SendCooldownSeconds) * time.Second,
			DailyLimit: c.Verify.DailyLimit,
		}, verify.NewMemoryCodeStore(), limiter),
		SmsVerifier: verify.NewSmsMockVerifier(verify.SmsMockConfig{
			FixedCode:  c.Verify.SmsMock.FixedCode,
			TTL:        time.Duration(c.Verify.CodeTTLSeconds) * time.Second,
			Cooldown:   time.Duration(c.Verify.SendCooldownSeconds) * time.Second,
			DailyLimit: c.Verify.DailyLimit,
		}, verify.NewMemoryCodeStore(), limiter),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		JwtAuth:    middleware.NewJwtAuthMiddleware(publicKey).Handle,
	}
}
