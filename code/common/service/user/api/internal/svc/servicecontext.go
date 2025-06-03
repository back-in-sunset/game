package svc

import (
	"user/api/internal/config"
	"user/rpc/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	UserRpc userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}

type Validater interface {
	SendCode(receiver string, code string) error
	Validate(code string) (bool, error)
}

type EmailValidater struct {
	Email string
}

type SmsValidater struct {
	Mobile string
}
