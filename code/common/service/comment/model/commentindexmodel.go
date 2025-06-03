package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentIndexModel = (*customCommentIndexModel)(nil)

type (
	// CommentIndexModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentIndexModel.
	CommentIndexModel interface {
		commentIndexModel
	}

	customCommentIndexModel struct {
		*defaultCommentIndexModel
		tableFn func(uint64) string
	}
)

// NewCommentIndexModel returns a model for the database table.
func NewCommentIndexModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentIndexModel {
	return &customCommentIndexModel{
		defaultCommentIndexModel: newCustomCommentIndexModel(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_index_%d`", shardingId&shardingBitmask)
		},
	}
}

func newCustomCommentIndexModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultCommentIndexModel {
	return &defaultCommentIndexModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`comment_index_0`",
	}
}

func (m *customCommentIndexModel) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	commentIndexIdKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(data.ObjId), id)
	commentIndexStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, commentIndexIdKey, commentIndexStateAttrsObjIdObjTypeKey)
	return err
}

func (m *customCommentIndexModel) FindOne(ctx context.Context, id uint64) (*CommentIndex, error) {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return nil, fmt.Errorf("objId is required in context")
	}
	commentIndexIdKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(objId), id)
	var resp CommentIndex
	err := m.QueryRowCtx(ctx, &resp, commentIndexIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentIndexRows, m.table)
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

func (m *customCommentIndexModel) FindOneByStateAttrsObjIdObjType(ctx context.Context, state uint64, attrs int64, objId uint64, objType uint64) (*CommentIndex, error) {
	commentIndexStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, state, attrs, objId, objType)
	var resp CommentIndex
	err := m.QueryRowIndexCtx(ctx, &resp, commentIndexStateAttrsObjIdObjTypeKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentIndexRows, m.table)
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

func (m *customCommentIndexModel) Insert(ctx context.Context, data *CommentIndex) (sql.Result, error) {
	commentIndexIdKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(data.ObjId), data.Id)
	commentIndexStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, commentIndexRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.ObjId, data.ObjType, data.MemberId, data.RootId, data.ReplyId, data.Floor, data.Count, data.RootCount, data.LikeCount, data.HateCount, data.State, data.Attrs)
	}, commentIndexIdKey, commentIndexStateAttrsObjIdObjTypeKey)
	return ret, err
}

func (m *customCommentIndexModel) Update(ctx context.Context, newData *CommentIndex) error {
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	commentIndexIdKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(data.ObjId), data.Id)
	commentIndexStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, commentIndexRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.RootId, newData.ReplyId, newData.Floor, newData.Count, newData.RootCount, newData.LikeCount, newData.HateCount, newData.State, newData.Attrs, newData.Id)
	}, commentIndexIdKey, commentIndexStateAttrsObjIdObjTypeKey)
	return err
}
