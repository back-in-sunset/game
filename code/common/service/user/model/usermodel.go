package model

import (
	"context"
	"database/sql"
	"fmt"

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
	userMobileKey := fmt.Sprintf("%s%v", cacheUserMobilePrefix, data.Mobile)
	userUserIDKey := fmt.Sprintf("%s%v", cacheUserUserIDPrefix, data.UserID)

	return m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("insert into %s (`user_id`, `name`, `gender`, `mobile`, `email`, `password`) values (?, ?, ?, ?, ?, ?)", m.tableName())
		return conn.ExecCtx(ctx, query, data.UserID, data.Name, data.Gender, data.Mobile, data.Email, data.Password)
	}, userMobileKey, userUserIDKey)
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
