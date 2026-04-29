package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserProfileModel = (*customUserProfileModel)(nil)

type (
	UserProfileModel interface {
		userProfileModel
		Upsert(ctx context.Context, data *UserProfile) error
	}

	customUserProfileModel struct {
		*defaultUserProfileModel
	}
)

func NewUserProfileModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserProfileModel {
	return &customUserProfileModel{
		defaultUserProfileModel: newUserProfileModel(conn, c, opts...),
	}
}

func (m *customUserProfileModel) Upsert(ctx context.Context, data *UserProfile) error {
	query := "insert into `user_profile` (`user_id`,`avatar`,`bio`,`birthday`,`location`,`extra`) values (?, ?, ?, ?, ?, ?) " +
		"on duplicate key update `avatar`=values(`avatar`),`bio`=values(`bio`),`birthday`=values(`birthday`),`location`=values(`location`),`extra`=values(`extra`)"
	cacheKey := fmt.Sprintf("%s%d", cacheUserProfileUserIDPrefix, data.UserID)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, data.UserID, data.Avatar, data.Bio, data.Birthday, data.Location, data.Extra)
	}, cacheKey)
	if isProfileTableNotExistsErr(err) {
		return nil
	}
	return err
}

func isProfileTableNotExistsErr(err error) bool {
	var mysqlErr *mysqlerr.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1146
}
