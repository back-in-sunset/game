package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		FindOneByEmail(ctx context.Context, email string) (*User, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

const (
	userTable            = "`user`"
	userMobileIndexTable = "`user_mobile_index`"

	cacheUserMobileIDPrefix = "cache:user:mobile:index:"
	cacheUserEmailPrefix    = "cache:user:email:"
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
}

func (m *customUserModel) FindOneByMobile(ctx context.Context, mobile string) (*User, error) {
	userID, err := m.findUserIDByMobileIndex(ctx, mobile)
	if err == nil {
		return m.defaultUserModel.FindOne(ctx, userID)
	}
	if err != ErrNotFound && !isTableNotExistsErr(err) {
		return nil, err
	}

	// 兼容索引表未部署场景。
	return m.defaultUserModel.FindOneByMobile(ctx, mobile)
}

func (m *customUserModel) FindOneByEmail(ctx context.Context, email string) (*User, error) {
	cacheKey := fmt.Sprintf("%s%v", cacheUserEmailPrefix, email)
	var resp User
	err := m.QueryRowCtx(ctx, &resp, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `email` = ? limit 1", userRows, userTable)
		return conn.QueryRowCtx(ctx, v, query, email)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	if data.UserID <= 0 {
		return nil, errors.New("user_id is required")
	}

	userMobileKey := fmt.Sprintf("%s%v", cacheUserMobilePrefix, data.Mobile)
	userMobileIDKey := fmt.Sprintf("%s%v", cacheUserMobileIDPrefix, data.Mobile)
	userUserIDKey := fmt.Sprintf("%s%v", cacheUserUserIDPrefix, data.UserID)

	userInsertSQL := fmt.Sprintf("insert into %s (`user_id`, `name`, `gender`, `mobile`, `email`, `password`) values (?, ?, ?, ?, ?, ?)", userTable)
	indexInsertSQL := fmt.Sprintf("insert into %s (`mobile`, `user_id`) values (?, ?)", userMobileIndexTable)

	var res sql.Result
	err := m.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		var execErr error
		res, execErr = s.ExecCtx(ctx, userInsertSQL, data.UserID, data.Name, data.Gender, data.Mobile, data.Email, data.Password)
		if execErr != nil {
			return execErr
		}

		_, execErr = s.ExecCtx(ctx, indexInsertSQL, data.Mobile, data.UserID)
		// 索引表未部署时兼容回退，保证注册可用。
		if execErr != nil && !isTableNotExistsErr(execErr) {
			return execErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = m.DelCacheCtx(ctx, userMobileKey, userMobileIDKey, userUserIDKey)
	return res, nil
}

func (m *customUserModel) FindOne(ctx context.Context, iD int64) (*User, error) {
	return m.defaultUserModel.FindOne(ctx, iD)
}

func (m *customUserModel) Update(ctx context.Context, data *User) error {
	oldData, err := m.defaultUserModel.FindOne(ctx, data.UserID)
	if err != nil {
		return err
	}

	userMobileKey := fmt.Sprintf("%s%v", cacheUserMobilePrefix, oldData.Mobile)
	newUserMobileKey := fmt.Sprintf("%s%v", cacheUserMobilePrefix, data.Mobile)
	oldMobileIDKey := fmt.Sprintf("%s%v", cacheUserMobileIDPrefix, oldData.Mobile)
	newMobileIDKey := fmt.Sprintf("%s%v", cacheUserMobileIDPrefix, data.Mobile)
	userUserIDKey := fmt.Sprintf("%s%v", cacheUserUserIDPrefix, oldData.UserID)

	userUpdateSQL := fmt.Sprintf("update %s set `name`=?,`gender`=?,`mobile`=?,`email`=?,`password`=? where `user_id`=?", userTable)
	indexDeleteSQL := fmt.Sprintf("delete from %s where `mobile`=?", userMobileIndexTable)
	indexInsertSQL := fmt.Sprintf("insert into %s (`mobile`, `user_id`) values (?, ?)", userMobileIndexTable)

	err = m.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		_, execErr := s.ExecCtx(ctx, userUpdateSQL, data.Name, data.Gender, data.Mobile, data.Email, data.Password, data.UserID)
		if execErr != nil {
			return execErr
		}

		if oldData.Mobile != data.Mobile {
			_, execErr = s.ExecCtx(ctx, indexDeleteSQL, oldData.Mobile)
			if execErr != nil && !isTableNotExistsErr(execErr) {
				return execErr
			}
			_, execErr = s.ExecCtx(ctx, indexInsertSQL, data.Mobile, data.UserID)
			if execErr != nil && !isTableNotExistsErr(execErr) {
				return execErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = m.DelCacheCtx(ctx, userMobileKey, newUserMobileKey, oldMobileIDKey, newMobileIDKey, userUserIDKey)
	return nil
}

func (m *customUserModel) Delete(ctx context.Context, iD int64) error {
	data, err := m.defaultUserModel.FindOne(ctx, iD)
	if err != nil {
		return err
	}

	userMobileKey := fmt.Sprintf("%s%v", cacheUserMobilePrefix, data.Mobile)
	userMobileIDKey := fmt.Sprintf("%s%v", cacheUserMobileIDPrefix, data.Mobile)
	userUserIDKey := fmt.Sprintf("%s%v", cacheUserUserIDPrefix, iD)

	userDeleteSQL := fmt.Sprintf("delete from %s where `user_id` = ?", userTable)
	indexDeleteSQL := fmt.Sprintf("delete from %s where `mobile` = ?", userMobileIndexTable)

	err = m.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		_, execErr := s.ExecCtx(ctx, userDeleteSQL, iD)
		if execErr != nil {
			return execErr
		}

		_, execErr = s.ExecCtx(ctx, indexDeleteSQL, data.Mobile)
		if execErr != nil && !isTableNotExistsErr(execErr) {
			return execErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = m.DelCacheCtx(ctx, userMobileKey, userMobileIDKey, userUserIDKey)
	return nil
}

func isTableNotExistsErr(err error) bool {
	var mysqlErr *mysqlerr.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1146
}

func (m *customUserModel) findUserIDByMobileIndex(ctx context.Context, mobile string) (int64, error) {
	type mobileIndex struct {
		UserID int64 `db:"user_id"`
	}
	cacheKey := fmt.Sprintf("%s%v", cacheUserMobileIDPrefix, mobile)
	var row mobileIndex
	err := m.QueryRowCtx(ctx, &row, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select `user_id` from %s where `mobile` = ? limit 1", userMobileIndexTable)
		return conn.QueryRowCtx(ctx, v, query, mobile)
	})
	switch err {
	case nil:
		return row.UserID, nil
	case sqlc.ErrNotFound:
		return 0, ErrNotFound
	default:
		return 0, err
	}
}
