package svc

import (
	"comment/model"
	"comment/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	CommentContentModel model.CommentContentModel
	CommentIndexModel   model.CommentIndexModel
	CommentSubjectModel model.CommentSubjectModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:              c,
		CommentContentModel: model.NewCommentContentModel(conn, c.CacheRedis),
		CommentIndexModel:   model.NewCommentIndexModel(conn, c.CacheRedis),
		CommentSubjectModel: model.NewCommentSubjectModel(conn, c.CacheRedis),
	}
}
