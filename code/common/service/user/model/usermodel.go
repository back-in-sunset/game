package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
}

func (m *customUserModel) FindOneByMobile(ctx context.Context, mobile string) (*User, error) {
	return m.defaultUserModel.FindOneByMobile(ctx, mobile)
}

func (m *customUserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	return m.defaultUserModel.Insert(ctx, data)
}
func (m *customUserModel) FindOne(ctx context.Context, iD int64) (*User, error) {
	return m.defaultUserModel.FindOne(ctx, iD)
}

func (m *customUserModel) Update(ctx context.Context, data *User) error {
	return m.defaultUserModel.Update(ctx, data)
}

func (m *customUserModel) Delete(ctx context.Context, iD int64) error {
	return m.defaultUserModel.Delete(ctx, iD)
}
