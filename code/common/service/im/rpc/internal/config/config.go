package config

import (
	baseconfig "im/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Auth      baseconfig.Auth
	Discovery baseconfig.Discovery
	Mysql     baseconfig.Mysql
	Redis     baseconfig.Redis
	Scope     baseconfig.ScopeDefaults
}
