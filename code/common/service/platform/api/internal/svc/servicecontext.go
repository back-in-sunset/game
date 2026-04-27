// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"platform/api/internal/config"
	"platform/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	Repo   *model.MySQLRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config: c,
		Repo:   model.NewMySQLRepository(conn, c.CacheRedis),
	}
}
