package config

import (
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

	UserRpc zrpc.RpcClientConf
}
