package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Auth struct {
		AccessSecret string
		AccessExpire int64
		// 非对称密钥路径
		PrivateKeyFile string
		PublicKeyFile  string
		Method         string
	}

	Mysql struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
	Salt       string
	Verify     struct {
		CodeTTLSeconds      int64
		SendCooldownSeconds int64
		DailyLimit          int64
		Email               struct {
			Host     string
			Port     int
			Username string
			Password string
			From     string
			Subject  string
		}
		SmsMock struct {
			FixedCode string
		}
	}

	UserRpc zrpc.RpcClientConf
}
