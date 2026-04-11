package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DeviceShadowModel = (*customDeviceShadowModel)(nil)

type (
	// DeviceShadowModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDeviceShadowModel.
	DeviceShadowModel interface {
		deviceShadowModel
		withSession(session sqlx.Session) DeviceShadowModel
	}

	customDeviceShadowModel struct {
		*defaultDeviceShadowModel
	}
)

// NewDeviceShadowModel returns a model for the database table.
func NewDeviceShadowModel(conn sqlx.SqlConn) DeviceShadowModel {
	return &customDeviceShadowModel{
		defaultDeviceShadowModel: newDeviceShadowModel(conn),
	}
}

func (m *customDeviceShadowModel) withSession(session sqlx.Session) DeviceShadowModel {
	return NewDeviceShadowModel(sqlx.NewSqlConnFromSession(session))
}
