package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentSubject0Model = (*customCommentSubject0Model)(nil)

var (
// cacheCommentSubject0IdPrefix = "cache:commentSubject0:id:"
)

type (
	// CommentSubject0Model is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentSubject0Model.
	CommentSubject0Model interface {
		commentSubject0Model
	}

	customCommentSubject0Model struct {
		*defaultCommentSubject0Model
		tableFn func(uint64) string
	}
)

// NewCommentSubject0Model returns a model for the database table.
func NewCommentSubject0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentSubject0Model {
	return &customCommentSubject0Model{
		defaultCommentSubject0Model: newCustomCommentSubject0Model(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
		},
	}
}

func newCustomCommentSubject0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultCommentSubject0Model {
	return &defaultCommentSubject0Model{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`comment_subject_0`",
	}
}

func (m *customCommentSubject0Model) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	commentSubject0IdKey := fmt.Sprintf("%s%v", cacheCommentSubject0IdPrefix, id)
	commentSubject0StateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubject0StateAttrsMemberIdPrefix, data.State, data.Attrs, data.MemberId)
	commentSubject0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubject0StateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, commentSubject0IdKey, commentSubject0StateAttrsMemberIdKey, commentSubject0StateAttrsObjIdObjTypeKey)
	return err
}

func (m *customCommentSubject0Model) FindOne(ctx context.Context, id uint64) (*CommentSubject0, error) {
	commentSubject0IdKey := fmt.Sprintf("%s%v", cacheCommentSubject0IdPrefix, id)
	var resp CommentSubject0
	err := m.QueryRowCtx(ctx, &resp, commentSubject0IdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentSubject0Rows, m.table)
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

func (m *customCommentSubject0Model) FindOneByStateAttrsMemberId(ctx context.Context, state uint64, attrs int64, memberId uint64) (*CommentSubject0, error) {
	commentSubject0StateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubject0StateAttrsMemberIdPrefix, state, attrs, memberId)
	var resp CommentSubject0
	err := m.QueryRowIndexCtx(ctx, &resp, commentSubject0StateAttrsMemberIdKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `member_id` = ? limit 1", commentSubject0Rows, m.table)
		if err := conn.QueryRowCtx(ctx, &resp, query, state, attrs, memberId); err != nil {
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

func (m *customCommentSubject0Model) FindOneByStateAttrsObjIdObjType(ctx context.Context, state uint64, attrs int64, objId uint64, objType uint64) (*CommentSubject0, error) {
	commentSubject0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubject0StateAttrsObjIdObjTypePrefix, state, attrs, objId, objType)
	var resp CommentSubject0
	err := m.QueryRowIndexCtx(ctx, &resp, commentSubject0StateAttrsObjIdObjTypeKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentSubject0Rows, m.table)
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

func (m *customCommentSubject0Model) Insert(ctx context.Context, data *CommentSubject0) (sql.Result, error) {
	commentSubject0IdKey := fmt.Sprintf("%s%v", cacheCommentSubject0IdPrefix, data.Id)
	commentSubject0StateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubject0StateAttrsMemberIdPrefix, data.State, data.Attrs, data.MemberId)
	commentSubject0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubject0StateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table, commentSubject0RowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.ObjId, data.ObjType, data.MemberId, data.Count, data.RootCount, data.AllCount, data.State, data.Attrs)
	}, commentSubject0IdKey, commentSubject0StateAttrsMemberIdKey, commentSubject0StateAttrsObjIdObjTypeKey)
	return ret, err
}

func (m *customCommentSubject0Model) Update(ctx context.Context, newData *CommentSubject0) error {
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	commentSubject0IdKey := fmt.Sprintf("%s%v", cacheCommentSubject0IdPrefix, data.Id)
	commentSubject0StateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubject0StateAttrsMemberIdPrefix, data.State, data.Attrs, data.MemberId)
	commentSubject0StateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubject0StateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, commentSubject0RowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.Count, newData.RootCount, newData.AllCount, newData.State, newData.Attrs, newData.Id)
	}, commentSubject0IdKey, commentSubject0StateAttrsMemberIdKey, commentSubject0StateAttrsObjIdObjTypeKey)
	return err
}
