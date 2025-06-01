package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentIndex0Model = (*customCommentIndex0Model)(nil)

type (
	// CommentIndex0Model is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentIndex0Model.
	CommentIndex0Model interface {
		commentIndex0Model
	}

	customCommentIndex0Model struct {
		*defaultCommentIndex0Model
		tableFn func(uint64) string
	}
)

// NewCommentIndex0Model returns a model for the database table.
func NewCommentIndex0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentIndex0Model {
	return &customCommentIndex0Model{
		defaultCommentIndex0Model: newCustomCommentIndex0Model(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_index_%d`", shardingId&shardingBitmask)
		},
	}
}

func newCustomCommentIndex0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultCommentIndex0Model {
	return &defaultCommentIndex0Model{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`comment_index_0`",
	}
}

func (m *customCommentIndex0Model) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	commentIndex0IdKey := fmt.Sprintf("%s%v", cacheCommentIndex0IdPrefix, id)
	commentIndex0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndex0StateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, commentIndex0IdKey, commentIndex0StateAttrsObjIdObjTypeKey)
	return err
}

func (m *customCommentIndex0Model) FindOne(ctx context.Context, id uint64) (*CommentIndex0, error) {
	commentIndex0IdKey := fmt.Sprintf("%s%v", cacheCommentIndex0IdPrefix, id)
	var resp CommentIndex0
	err := m.QueryRowCtx(ctx, &resp, commentIndex0IdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentIndex0Rows, m.table)
		return conn.QueryRowCtx(ctx, v, query, id)
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

func (m *customCommentIndex0Model) FindOneByStateAttrsObjIdObjType(ctx context.Context, state uint64, attrs int64, objId uint64, objType uint64) (*CommentIndex0, error) {
	commentIndex0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndex0StateAttrsObjIdObjTypePrefix, state, attrs, objId, objType)
	var resp CommentIndex0
	err := m.QueryRowIndexCtx(ctx, &resp, commentIndex0StateAttrsObjIdObjTypeKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentIndex0Rows, m.table)
		if err := conn.QueryRowCtx(ctx, &resp, query, state, attrs, objId, objType); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customCommentIndex0Model) Insert(ctx context.Context, data *CommentIndex0) (sql.Result, error) {
	commentIndex0IdKey := fmt.Sprintf("%s%v", cacheCommentIndex0IdPrefix, data.Id)
	commentIndex0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndex0StateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, commentIndex0RowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.ObjId, data.ObjType, data.MemberId, data.RootId, data.ReplyId, data.Floor, data.Count, data.RootCount, data.LikeCount, data.HateCount, data.State, data.Attrs)
	}, commentIndex0IdKey, commentIndex0StateAttrsObjIdObjTypeKey)
	return ret, err
}

func (m *customCommentIndex0Model) Update(ctx context.Context, newData *CommentIndex0) error {
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	commentIndex0IdKey := fmt.Sprintf("%s%v", cacheCommentIndex0IdPrefix, data.Id)
	commentIndex0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndex0StateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, commentIndex0RowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.RootId, newData.ReplyId, newData.Floor, newData.Count, newData.RootCount, newData.LikeCount, newData.HateCount, newData.State, newData.Attrs, newData.Id)
	}, commentIndex0IdKey, commentIndex0StateAttrsObjIdObjTypeKey)
	return err
}
