package svc

import (
	"comment/model"
	"comment/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/singleflight"
)

type ServiceContext struct {
	Config            config.Config
	SingleFlightGroup singleflight.Group
	BizRedis          *redis.Redis

	CommentContentModel model.CommentContentModel
	CommentIndexModel   model.CommentIndexModel
	CommentSubjectModel model.CommentSubjectModel
	CommentModel        model.CommentModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	rds, err := redis.NewRedis(redis.RedisConf{
		Host: c.BizRedis.Host,
		Pass: c.BizRedis.Pass,
		Type: c.BizRedis.Type,
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:              c,
		BizRedis:            rds,
		CommentContentModel: model.NewCommentContentModel(conn, c.CacheRedis),
		CommentIndexModel:   model.NewCommentIndexModel(conn, c.CacheRedis),
		CommentSubjectModel: model.NewCommentSubjectModel(conn, c.CacheRedis),
		CommentModel:        model.NewCommentModel(conn, c.CacheRedis),
	}
}
