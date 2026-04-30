package svc

import (
	"history/internal/historycache"
	"history/model"
	"history/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	HistoryModel *historycache.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	var rds *redis.Redis
	if len(c.CacheRedis) > 0 {
		redisClient, err := redis.NewRedis(c.CacheRedis[0].RedisConf)
		if err != nil {
			panic(err)
		}
		rds = redisClient
	}

	return &ServiceContext{
		Config:       c,
		HistoryModel: historycache.NewManager(model.NewHistoryModel(conn), rds, c.HistoryCache),
	}
}

func (s *ServiceContext) Start() {
	if s.HistoryModel != nil {
		s.HistoryModel.Start()
	}
}

func (s *ServiceContext) Stop() {
	if s.HistoryModel != nil {
		s.HistoryModel.Stop()
	}
}
