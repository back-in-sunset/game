// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"
	"crypto/rsa"
	"os"
	"platform/api/internal/config"
	"platform/api/internal/middleware"
	"platform/internal/domain"
	"platform/model"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type Repository interface {
	SaveTenant(ctx context.Context, tenant *domain.Tenant) error
	UpdateTenant(ctx context.Context, tenant *domain.Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error)
	SaveProject(ctx context.Context, project *domain.Project) error
	ListProjectsByTenantID(ctx context.Context, tenantID string) ([]*domain.Project, error)
	SaveEnvironment(ctx context.Context, environment *domain.Environment) error
	ListEnvironmentsByProjectID(ctx context.Context, projectID string) ([]*domain.Environment, error)
	ListTenantsByMemberID(ctx context.Context, memberID string) ([]*domain.Tenant, error)
}

type ServiceContext struct {
	Config config.Config
	Repo   Repository

	JwtAuth   rest.Middleware
	PublicKey *rsa.PublicKey
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	publicKeyData, err := os.ReadFile(c.Auth.PublicKeyFile)
	if err != nil {
		panic(err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:    c,
		Repo:      model.NewMySQLRepository(conn, c.CacheRedis),
		PublicKey: publicKey,
		JwtAuth:   middleware.NewJwtAuthMiddleware(publicKey).Handle,
	}
}
