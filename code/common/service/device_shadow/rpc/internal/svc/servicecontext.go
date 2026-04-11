package svc

import (
	"game/server/device_shadow/rpc/internal/config"
	"game/server/device_shadow/rpc/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	ShadowModel model.DeviceShadowModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:      c,
		ShadowModel: model.NewDeviceShadowModel(conn),
	}
}
