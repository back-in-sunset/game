package model

import (
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
		// Insert(ctx context.Context, data *CommentSubject) (sql.Result, error)
		// FindOne(ctx context.Context, objID, id uint64) (*CommentSubject, error)
		// FindOneByStateMemberID(ctx context.Context, state uint64, objID, memberID uint64) (*CommentSubject, error)
		// FindOneByStateObjIDObjType(ctx context.Context, state uint64, objID uint64, objType uint64) (*CommentSubject, error)
		// Update(ctx context.Context, data *CommentSubject) error
		// Delete(ctx context.Context, objID, id uint64) error
	}

	customCommentSubjectModel struct {
		sqlc.CachedConn
		*defaultCommentSubjectModel
		// tableFn func(uint64) string
		table string
	}

	newCustomCommentSubjectModelFunc func(shardingID int64) *customCommentSubjectModel
)

// // NewCommentSubjectModel returns a model for the database table.
// func NewCommentSubjectModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentSubjectModel {
// 	return &customCommentSubjectModel{
// 		defaultCommentSubjectModel: newCommentSubjectModel(conn, c, opts...),
// 		tableFn: func(shardingId uint64) string {
// 			// Use the last 8 bits of the shardingId for determining the table suffix.
// 			//  const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
// 			return fmt.Sprintf("`comment_subject_%d`", shardingId&0xFF) // Adjust the bitmask as needed
// 		},
// 	}
// }

func newCustomCommentSubjectModel(conn sqlx.SqlConn, c cache.CacheConf, shardingID int64, opts ...cache.Option) *customCommentSubjectModel {
	table := fmt.Sprintf("`comment_subject_%d`", shardingID&0xFF)
	commentSubjectModel := newCommentSubjectModel(conn, c, opts...)
	commentSubjectModel.table = table
	return &customCommentSubjectModel{
		defaultCommentSubjectModel: commentSubjectModel,
		table:                      table,
		CachedConn:                 sqlc.NewConn(conn, c, opts...),
	}
}

// func (m newCustomCommentSubjectFunc) Delete(ctx context.Context, objID, id uint64) error {
// 	return m(objID).Delete(ctx, objID, id)
// }

// func (m *customCommentSubjectModel) Delete(ctx context.Context, objID, id uint64) error {
// 	data, err := m.FindOne(ctx, objID, id)
// 	if err != nil {
// 		return err
// 	}

// 	commentSubjectIDKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), id)
// 	commentSubjectStateAttrsMemberIDKey := fmt.Sprintf("%s%s:%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(data.ObjId), data.State, data.MemberId)
// 	commentSubjectStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
// 	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
// 		query := fmt.Sprintf("delete from %s where `id` = ?", m.tableFn(data.ObjId))
// 		return conn.ExecCtx(ctx, query, id)
// 	}, commentSubjectIDKey, commentSubjectStateAttrsMemberIDKey, commentSubjectStateAttrsObjIDObjTypeKey)
// 	return err
// }

// func (m *customCommentSubjectModel) FindOne(ctx context.Context, objID, id uint64) (*CommentSubject, error) {
// 	commentSubjectIDKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objID), id)
// 	var resp CommentSubject
// 	err := m.QueryRowCtx(ctx, &resp, commentSubjectIDKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
// 		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentSubjectRows, m.tableFn(objID))
// 		return conn.QueryRowCtx(ctx, v, query, id)
// 	})
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	case sqlc.ErrNotFound:
// 		return nil, ErrNotFound
// 	default:
// 		return nil, err
// 	}
// }

// func (m *customCommentSubjectModel) FindOneByStateAttrsMemberID(ctx context.Context, state uint64, attrs int64, objID, memberID uint64) (*CommentSubject, error) {
// 	commentSubjectStateAttrsMemberIDKey := fmt.Sprintf("%s%s%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(objID), state, memberID)
// 	var resp CommentSubject
// 	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateAttrsMemberIDKey, func(primary any) string {
// 		return fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objID), primary)
// 	}, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
// 		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `member_id` = ? limit 1", commentSubjectRows, m.tableFn(objID))
// 		if err := conn.QueryRowCtx(ctx, &resp, query, state, attrs, memberID); err != nil {
// 			return nil, err
// 		}
// 		return resp.ID, nil
// 	}, m.queryPrimary)
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	case sqlc.ErrNotFound:
// 		return nil, ErrNotFound
// 	default:
// 		return nil, err
// 	}
// }

// func (m *customCommentSubjectModel) FindOneByStateAttrsObjIDObjType(ctx context.Context, state, objID uint64, objType uint64) (*CommentSubject, error) {
// 	commentSubjectStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%s%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, m.tableFn(objID), state, objID, objType)
// 	var resp CommentSubject

// 	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateAttrsObjIDObjTypeKey, func(primary any) string {
// 		return fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(objID), primary)

// 	}, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
// 		query := fmt.Sprintf("select %s from %s where `obj_id` = ? and `obj_type` = ? limit 1", commentSubjectRows, m.tableFn(objID))
// 		if state > 0 {
// 			query = fmt.Sprintf("select %s from %s where `state` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentSubjectRows, m.tableFn(objID))
// 			if err := conn.QueryRowCtx(ctx, &resp, query, state, objID, objType); err != nil {
// 				return nil, err
// 			}
// 		} else {
// 			if err := conn.QueryRowCtx(ctx, &resp, query, objID, objType); err != nil {
// 				return nil, err
// 			}
// 		}

// 		return resp.ID, nil
// 	}, m.queryPrimary)
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	case sqlc.ErrNotFound:
// 		return nil, ErrNotFound
// 	default:
// 		return nil, err
// 	}
// }

// func (m *customCommentSubjectModel) Insert(ctx context.Context, data *CommentSubject) (sql.Result, error) {
// 	commentSubjectIDKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), data.ID)
// 	commentSubjectStateAttrsMemberIDKey := fmt.Sprintf("%s%s%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(data.ObjId), data.State, data.MemberId)
// 	commentSubjectStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%s%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, m.tableFn(data.ObjId), data.State, data.ObjId, data.ObjType)
// 	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
// 		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.tableFn(data.ObjId), commentSubjectRowsExpectAutoSet)
// 		return conn.ExecCtx(ctx, query, data.ObjId, data.ObjType, data.MemberId, data.Count, data.RootCount, data.AllCount, data.State, data.Attrs)
// 	}, commentSubjectIDKey, commentSubjectStateAttrsMemberIDKey, commentSubjectStateAttrsObjIDObjTypeKey)
// 	return ret, err
// }

// func (m *customCommentSubjectModel) Update(ctx context.Context, newData *CommentSubject) error {
// 	if newData.ObjId == 0 {
// 		return fmt.Errorf("objId is required in newData")
// 	}
// 	data, err := m.FindOne(ctx, newData.ObjId, newData.ID)
// 	if err != nil {
// 		return err
// 	}

// 	commentSubjectIDKey := fmt.Sprintf("%s%s%v", cacheCommentSubjectIdPrefix, m.tableFn(data.ObjId), data.ID)
// 	commentSubjectStateAttrsMemberIDKey := fmt.Sprintf("%s%s:%v:%v", cacheCommentSubjectStateMemberIdPrefix, m.tableFn(data.ObjId), data.State, data.MemberId)
// 	commentSubjectStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentSubjectStateObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
// 	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
// 		query := fmt.Sprintf("update %s set %s where `id` = ?", m.tableFn(data.ObjId), commentSubjectRowsWithPlaceHolder)
// 		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.Count, newData.RootCount, newData.AllCount, newData.State, newData.Attrs, newData.ID)
// 	}, commentSubjectIDKey, commentSubjectStateAttrsMemberIDKey, commentSubjectStateAttrsObjIDObjTypeKey)
// 	return err
// }

// func (m *customCommentSubjectModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
// 	objID, ok := ctx.Value(objIDStruct).(uint64)
// 	if !ok || objID == 0 {
// 		return fmt.Errorf("objId is required in context")
// 	}
// 	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentSubjectRows, m.tableFn(objID))
// 	return conn.QueryRowCtx(ctx, v, query, primary)
// }

// func (m *customCommentSubjectModel) FindOneByStateMemberID(ctx context.Context, state uint64, objID, memberID uint64) (*CommentSubject, error) {
// 	commentSubjectStateMemberIDKey := fmt.Sprintf("%s%v:%v:%v", cacheCommentSubjectStateMemberIdPrefix, state, objID, memberID)
// 	var resp CommentSubject
// 	err := m.QueryRowIndexCtx(ctx, &resp, commentSubjectStateMemberIDKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
// 		query := fmt.Sprintf("select %s from %s where `state` = ? and `member_id` = ? limit 1", commentSubjectRows, m.tableFn(objID))
// 		if err := conn.QueryRowCtx(ctx, &resp, query, state, memberID); err != nil {
// 			return nil, err
// 		}
// 		return resp.ID, nil
// 	}, m.queryPrimary)
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	case sqlc.ErrNotFound:
// 		return nil, ErrNotFound
// 	default:
// 		return nil, err
// 	}
// }
