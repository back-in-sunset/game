package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Identity struct {
		Timeout   int
		PoolSize  int
		BatchSize int
	}

	Mysql struct {
		DataSource string
	}

	CacheRedis cache.CacheConf

	Salt string

	ScyllaDB struct {
		Hosts             []string
		Keyspace          string
		Timeout           int
		NumConns          int
		ReconnectInterval int
	}
}
