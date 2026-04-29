package svc

import (
	"history/model"
	"history/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	HistoryModel model.HistoryModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:       c,
		HistoryModel: model.NewHistoryModel(sqlx.NewMysql(c.Mysql.DataSource)),
	}
}
