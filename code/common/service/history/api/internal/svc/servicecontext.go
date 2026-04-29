package svc

import (
	"history/api/internal/config"
	"history/model"
	"history/rpc/historyclient"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	HistoryRpc historyclient.History
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.Mysql.DataSource != "" {
		return &ServiceContext{
			Config:     c,
			HistoryRpc: historyclient.NewLocalHistory(model.NewHistoryModel(sqlx.NewMysql(c.Mysql.DataSource))),
		}
	}
	return &ServiceContext{
		Config:     c,
		HistoryRpc: historyclient.NewHistory(zrpc.MustNewClient(c.HistoryRPC)),
	}
}
