package svc

import (
	"comment/model"
	"comment/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	CommentModel model.CommentModel // Assuming CommentModel is defined elsewhere
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:       c,
		CommentModel: model.NewCommentModel(conn, c.CacheRedis), // Assuming CacheRedis is defined in config
	}
}
