package svc

import (
	"comment/api/internal/config"
	"comment/rpc/commentclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	CommentRpc commentclient.Comment
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		CommentRpc: commentclient.NewComment(zrpc.MustNewClient(c.CommentRpc)),
	}
}
