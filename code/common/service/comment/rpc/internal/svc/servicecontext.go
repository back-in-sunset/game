package svc

import (
	"comment/model"
	"comment/rpc/internal/config"

	"golang.org/x/sync/singleflight"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config            config.Config
	CommentModel      model.CommentModel // Assuming CommentModel is defined elsewhere
	SignleFlightGroup singleflight.Group
	BizRedis          *redis.Redis
}

// NewServiceContext 创建服务上下文
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
		Config:       c,
		CommentModel: model.NewCommentModel(conn, c.CacheRedis), // Assuming CacheRedis is defined in config
		BizRedis:     rds,
	}
}
