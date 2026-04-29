package svc

import (
	"user/model"
	"user/rpc/internal/config"
	"user/utils/idx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config config.Config
	IdxGen *idx.Snowflake

	UserModel        model.UserModel
	UserProfileModel model.UserProfileModel
}

// NewServiceContext 创建 ServiceContext
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:           c,
		UserModel:        model.NewUserModel(conn, c.CacheRedis),
		UserProfileModel: model.NewUserProfileModel(conn, c.CacheRedis),
		// UserModel: newCQLUserModel(c),
		IdxGen: newIdxGenerator(c),
	}
}
