package svc

import (
	"friend/internal/friendcache"
	"friend/internal/repository"
	mysqlrepo "friend/internal/repository/mysql"
	"friend/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	FriendStore repository.FriendStore
	Cache       *friendcache.Cache
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

	store := newFriendStore(c, conn)
	cache := friendcache.NewCache(rds, c.FriendCache.ListTTLSeconds, c.FriendCache.CheckTTLSeconds)

	return &ServiceContext{
		Config:      c,
		FriendStore: store,
		Cache:       cache,
	}
}

func newFriendStore(c config.Config, conn sqlx.SqlConn) repository.FriendStore {
	switch c.FriendStore.Backend {
	case "scylla":
		panic("scylla backend not yet implemented")
	default:
		return mysqlrepo.NewFriendStore(conn)
	}
}
