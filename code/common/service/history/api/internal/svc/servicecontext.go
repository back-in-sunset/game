package svc

import (
	"history/api/internal/config"
	"history/internal/historycache"
	"history/model"
	"history/rpc/historyclient"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	HistoryRpc historyclient.History
	localCache *historycache.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.Mysql.DataSource != "" {
		var rds *redis.Redis
		if len(c.CacheRedis) > 0 {
			redisClient, err := redis.NewRedis(c.CacheRedis[0].RedisConf)
			if err != nil {
				panic(err)
			}
			rds = redisClient
		}
		manager := historycache.NewManager(model.NewHistoryModel(sqlx.NewMysql(c.Mysql.DataSource)), rds, c.HistoryCache)
		return &ServiceContext{
			Config:     c,
			HistoryRpc: historyclient.NewLocalHistory(manager),
			localCache: manager,
		}
	}
	return &ServiceContext{
		Config:     c,
		HistoryRpc: historyclient.NewHistory(zrpc.MustNewClient(c.HistoryRPC)),
	}
}

func (s *ServiceContext) Start() {
	if s.localCache != nil {
		s.localCache.Start()
	}
}

func (s *ServiceContext) Stop() {
	if s.localCache != nil {
		s.localCache.Stop()
	}
}
