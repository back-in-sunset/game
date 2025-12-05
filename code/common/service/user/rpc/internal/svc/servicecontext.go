package svc

import (
	"user/model"
	"user/rpc/internal/config"
	"user/utils/idx"
)

type ServiceContext struct {
	Config config.Config

	UserModel model.UserModel

	IdxGen *idx.Snowflake
}

// NewServiceContext 创建 ServiceContext
func NewServiceContext(c config.Config) *ServiceContext {
	// conn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config: c,
		// UserModel: model.NewUserModel(conn, c.CacheRedis),
		UserModel: newCQLUserModel(c),
		IdxGen:    newIdxGenerator(c),
	}
}
