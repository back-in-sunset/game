package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentSubjectModel = (*customCommentSubjectModel)(nil)

var (
// cacheCommentSubjectShardingIdIdPrefix = "cache:commentSubject:id:"
)

type (
	// CommentSubjectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentSubjectModel.
	CommentSubjectModel interface {
		commentSubjectModel
	}

	customCommentSubjectModel struct {
		*defaultCommentSubjectModel
		tableFn func(uint64) string
	}
)

// NewCommentSubjectModel returns a model for the database table.
func NewCommentSubjectModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentSubjectModel {
	return &customCommentSubjectModel{
		defaultCommentSubjectModel: newCustomCommentSubjectModel(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
		},
	}
}

func newCustomCommentSubjectModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultCommentSubjectModel {
	return &defaultCommentSubjectModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`comment_subject_0`",
	}
}

func (m *customCommentSubjectModel) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	commentSubjectIdKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), id)
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubjectStateAttrsMemberIdPrefix, data.State, data.Attrs, data.MemberId)
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, commentSubjectIdKey, commentSubjectStateAttrsMemberIdKey, commentSubjectStateAttrsObjIdObjTypeKey)
	return err
}

func (m *customCommentSubjectModel) FindOne(ctx context.Context, id uint64) (*CommentSubject, error) {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return nil, fmt.Errorf("objId is required in context")
	}
	commentSubjectIdKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objId), id)
	var resp CommentSubject
	err := m.QueryRowCtx(ctx, &resp, commentSubjectIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentSubjectRows, m.table)
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

func (m *customCommentSubjectModel) FindOneByStateAttrsMemberId(ctx context.Context, state uint64, attrs int64, memberId uint64) (*CommentSubject, error) {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return nil, fmt.Errorf("objId is required in context")
	}
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%s%v:%v:%v", cacheCommentSubjectStateAttrsMemberIdPrefix, m.tableFn(objId), state, attrs, memberId)
	var resp CommentSubject
	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateAttrsMemberIdKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `member_id` = ? limit 1", commentSubjectRows, m.table)
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

func (m *customCommentSubjectModel) FindOneByStateAttrsObjIdObjType(ctx context.Context, state uint64, attrs int64, objId uint64, objType uint64) (*CommentSubject, error) {
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%s%v:%v:%v:%v", cacheCommentSubjectStateAttrsObjIdObjTypePrefix, m.tableFn(objId), state, attrs, objId, objType)
	var resp CommentSubject
	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateAttrsObjIdObjTypeKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentSubjectRows, m.table)
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

func (m *customCommentSubjectModel) Insert(ctx context.Context, data *CommentSubject) (sql.Result, error) {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return nil, fmt.Errorf("objId is required in context")
	}
	commentSubjectIdKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objId), data.Id)
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubjectStateAttrsMemberIdPrefix, data.State, data.Attrs, data.MemberId)
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table, commentSubjectRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.ObjId, data.ObjType, data.MemberId, data.Count, data.RootCount, data.AllCount, data.State, data.Attrs)
	}, commentSubjectIdKey, commentSubjectStateAttrsMemberIdKey, commentSubjectStateAttrsObjIdObjTypeKey)
	return ret, err
}

func (m *customCommentSubjectModel) Update(ctx context.Context, newData *CommentSubject) error {
	if newData.ObjId == 0 {
		return fmt.Errorf("objId is required in newData")
	}
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	commentSubjectIdKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), data.Id)
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubjectStateAttrsMemberIdPrefix, data.State, data.Attrs, data.MemberId)
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, commentSubjectRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.Count, newData.RootCount, newData.AllCount, newData.State, newData.Attrs, newData.Id)
	}, commentSubjectIdKey, commentSubjectStateAttrsMemberIdKey, commentSubjectStateAttrsObjIdObjTypeKey)
	return err
}
