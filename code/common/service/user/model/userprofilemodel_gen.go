package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	userProfileFieldNames          = builder.RawFieldNames(&UserProfile{})
	userProfileRows                = strings.Join(userProfileFieldNames, ",")
	userProfileRowsExpectAutoSet   = strings.Join(stringx.Remove(userProfileFieldNames, "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	userProfileRowsWithPlaceHolder = strings.Join(stringx.Remove(userProfileFieldNames, "`user_id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"

	cacheUserProfileUserIDPrefix = "cache:userProfile:userID:"
)

type (
	userProfileModel interface {
		Insert(ctx context.Context, data *UserProfile) (sql.Result, error)
		FindOne(ctx context.Context, userID int64) (*UserProfile, error)
		Update(ctx context.Context, data *UserProfile) error
		Delete(ctx context.Context, userID int64) error
	}

	defaultUserProfileModel struct {
		sqlc.CachedConn
		table string
	}

	UserProfile struct {
		UserID    int64          `db:"user_id"`
		Avatar    string         `db:"avatar"`
		Bio       string         `db:"bio"`
		Birthday  sql.NullTime   `db:"birthday"`
		Location  string         `db:"location"`
		Extra     sql.NullString `db:"extra"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
	}
)

func newUserProfileModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultUserProfileModel {
	return &defaultUserProfileModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`user_profile`",
	}
}

func (m *defaultUserProfileModel) Delete(ctx context.Context, userID int64) error {
	userProfileUserIDKey := fmt.Sprintf("%s%v", cacheUserProfileUserIDPrefix, userID)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `user_id` = ?", m.table)
		return conn.ExecCtx(ctx, query, userID)
	}, userProfileUserIDKey)
	return err
}

func (m *defaultUserProfileModel) FindOne(ctx context.Context, userID int64) (*UserProfile, error) {
	userProfileUserIDKey := fmt.Sprintf("%s%v", cacheUserProfileUserIDPrefix, userID)
	var resp UserProfile
	err := m.QueryRowCtx(ctx, &resp, userProfileUserIDKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `user_id` = ? limit 1", userProfileRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, userID)
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

func (m *defaultUserProfileModel) Insert(ctx context.Context, data *UserProfile) (sql.Result, error) {
	userProfileUserIDKey := fmt.Sprintf("%s%v", cacheUserProfileUserIDPrefix, data.UserID)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, userProfileRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.UserID, data.Avatar, data.Bio, data.Birthday, data.Location, data.Extra)
	}, userProfileUserIDKey)
	return ret, err
}

func (m *defaultUserProfileModel) Update(ctx context.Context, data *UserProfile) error {
	userProfileUserIDKey := fmt.Sprintf("%s%v", cacheUserProfileUserIDPrefix, data.UserID)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `user_id` = ?", m.table, userProfileRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, data.Avatar, data.Bio, data.Birthday, data.Location, data.Extra, data.UserID)
	}, userProfileUserIDKey)
	return err
}
