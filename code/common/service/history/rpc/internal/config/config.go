package config

import (
	"history/internal/historycache"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	CacheRedis   cache.CacheConf
	HistoryCache historycache.Config
}
