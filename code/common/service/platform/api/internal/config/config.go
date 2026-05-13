// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		PublicKeyFile  string
		PrivateKeyFile string
	}
	LiveKit struct {
		ApiKey    string
		ApiSecret string
	}
	Mysql struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
}
