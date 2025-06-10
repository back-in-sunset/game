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
		defaultCommentSubjectModel: newCommentSubjectModel(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
		},
	}
}

func newCustomCommentSubjectModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *customCommentSubjectModel {
	return &customCommentSubjectModel{
		defaultCommentSubjectModel: newCommentSubjectModel(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
		},
	}
}

func (m *customCommentSubjectModel) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	commentSubjectIdKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), id)
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%s:%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(data.ObjId), data.State, data.MemberId)
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.tableFn(data.ObjId))
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
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentSubjectRows, m.tableFn(objId))
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
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%s%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(objId), state, memberId)
	var resp CommentSubject
	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateAttrsMemberIdKey, func(primary any) string {
		return fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objId), primary)
	}, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `member_id` = ? limit 1", commentSubjectRows, m.tableFn(objId))
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

func (m *customCommentSubjectModel) FindOneByStateAttrsObjIdObjType(ctx context.Context, state, objId uint64, objType uint64) (*CommentSubject, error) {
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%s%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, m.tableFn(objId), state, objId, objType)
	var resp CommentSubject

	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateAttrsObjIdObjTypeKey, func(primary any) string {
		return fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objId), primary)

	}, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `obj_id` = ? and `obj_type` = ? limit 1", commentSubjectRows, m.tableFn(objId))
		if state > 0 {
			query = fmt.Sprintf("select %s from %s where `state` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentSubjectRows, m.tableFn(objId))
			if err := conn.QueryRowCtx(ctx, &resp, query, state, objId, objType); err != nil {
				return nil, err
			}
		} else {
			if err := conn.QueryRowCtx(ctx, &resp, query, objId, objType); err != nil {
				return nil, err
			}
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
	commentSubjectIdKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), data.Id)
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%s%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(data.ObjId), data.State, data.MemberId)
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%s%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, m.tableFn(data.ObjId), data.State, data.ObjId, data.ObjType)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.tableFn(data.ObjId), commentSubjectRowsExpectAutoSet)
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
	commentSubjectStateAttrsMemberIdKey := fmt.Sprintf("%s%s:%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(data.ObjId), data.State, data.MemberId)
	commentSubjectStateAttrsObjIdObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.tableFn(data.ObjId), commentSubjectRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.Count, newData.RootCount, newData.AllCount, newData.State, newData.Attrs, newData.Id)
	}, commentSubjectIdKey, commentSubjectStateAttrsMemberIdKey, commentSubjectStateAttrsObjIdObjTypeKey)
	return err
}

func (m *customCommentSubjectModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return fmt.Errorf("objId is required in context")
	}
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentSubjectRows, m.tableFn(objId))
	return conn.QueryRowCtx(ctx, v, query, primary)
}

// func (m *customCommentSubjectModel) formatPrimary(primary any) string {
// 	return fmt.Sprintf("%s%v", cacheCommentSubjectIdPrefix, primary)
// }
