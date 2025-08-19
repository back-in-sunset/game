package model

import (
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
		// Insert(ctx context.Context, data *CommentIndex) (sql.Result, error)
		// FindOne(ctx context.Context, objID, id int64) (*CommentIndex, error)
		// FindOneByStateAttrsObjIDObjType(ctx context.Context, state int64, attrs int64, objID int64, objType int64) (*CommentIndex, error)
		// Update(ctx context.Context, data *CommentIndex) error
		// Delete(ctx context.Context, objID, id int64) error
	}

	customCommentIndexModel struct {
		*defaultCommentIndexModel
		sqlc.CachedConn
		// tableFn func(int64) string
		table string
	}

	newCustomCommentIndexModelFunc func(shardingID int64) *customCommentIndexModel
)

func newCustomCommentIndexModel(conn sqlx.SqlConn, c cache.CacheConf, shardingID int64, opts ...cache.Option) *customCommentIndexModel {
	table := fmt.Sprintf("`comment_index_%d`", shardingID&0xFF)
	defaultCommentIndexModel := newCommentIndexModel(conn, c, opts...)
	defaultCommentIndexModel.table = table
	return &customCommentIndexModel{
		CachedConn:               sqlc.NewConn(conn, c, opts...),
		defaultCommentIndexModel: defaultCommentIndexModel,
		table:                    table,
		// CachedConn: sqlc.NewConn(conn, c, opts...),
		// tableFn: func(shardingId int64) string {
		// 	// Use the last 8 bits of the shardingId for determining the table suffix.
		// 	const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
		// 	return fmt.Sprintf("`comment_index_%d`", shardingId&shardingBitmask)
		// },
	}
}

// func (m *customCommentIndexModel) Delete(ctx context.Context, objID, id int64) error {
// 	data, err := m.FindOne(ctx, objID, id)
// 	if err != nil {
// 		return err
// 	}

// 	commentIndexIDKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(data.ObjId), id)
// 	commentIndexStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
// 	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
// 		query := fmt.Sprintf("delete from %s where `id` = ?", m.tableFn(data.ObjId))
// 		return conn.ExecCtx(ctx, query, id)
// 	}, commentIndexIDKey, commentIndexStateAttrsObjIDObjTypeKey)
// 	return err
// }

// func (m *customCommentIndexModel) FindOne(ctx context.Context, objID, id int64) (*CommentIndex, error) {
// 	commentIndexIDKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(objID), id)
// 	var resp CommentIndex
// 	err := m.QueryRowCtx(ctx, &resp, commentIndexIDKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
// 		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", commentIndexRows, m.tableFn(objID))
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

// func (m *customCommentIndexModel) FindOneByStateAttrsObjIDObjType(ctx context.Context, state int64, attrs int64, objID int64, objType int64) (*CommentIndex, error) {
// 	commentIndexStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, state, attrs, objID, objType)
// 	var resp CommentIndex
// 	err := m.QueryRowIndexCtx(ctx, &resp, commentIndexStateAttrsObjIDObjTypeKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
// 		query := fmt.Sprintf("select %s from %s where `state` = ? and `attrs` = ? and `obj_id` = ? and `obj_type` = ? limit 1", commentIndexRows, m.tableFn(objID))
// 		if err := conn.QueryRowCtx(ctx, &resp, query, state, attrs, objID, objType); err != nil {
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

// func (m *customCommentIndexModel) Insert(ctx context.Context, data *CommentIndex) (sql.Result, error) {
// 	commentIndexIDKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(data.ObjId), data.ID)
// 	commentIndexStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
// 	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
// 		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.tableFn(data.ObjId), commentIndexRowsExpectAutoSet)
// 		return conn.ExecCtx(ctx, query, data.ObjId, data.ObjType, data.MemberId, data.RootId, data.ReplyId, data.Floor, data.Count, data.RootCount, data.LikeCount, data.HateCount, data.State, data.Attrs)
// 	}, commentIndexIDKey, commentIndexStateAttrsObjIDObjTypeKey)
// 	return ret, err
// }

// func (m *customCommentIndexModel) Update(ctx context.Context, newData *CommentIndex) error {
// 	data, err := m.FindOne(ctx, newData.ObjId, newData.ID)
// 	if err != nil {
// 		return err
// 	}

// 	commentIndexIDKey := fmt.Sprintf("%s%s%v", cacheCommentIndexIdPrefix, m.tableFn(data.ObjId), data.ID)
// 	commentIndexStateAttrsObjIDObjTypeKey := fmt.Sprintf("%s%v:%v:%v:%v", cacheCommentIndexStateAttrsObjIdObjTypePrefix, data.State, data.Attrs, data.ObjId, data.ObjType)
// 	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
// 		query := fmt.Sprintf("update %s set %s where `id` = ?", m.tableFn(data.ObjId), commentIndexRowsWithPlaceHolder)
// 		return conn.ExecCtx(ctx, query, newData.ObjId, newData.ObjType, newData.MemberId, newData.RootId, newData.ReplyId, newData.Floor, newData.Count, newData.RootCount, newData.LikeCount, newData.HateCount, newData.State, newData.Attrs, newData.ID)
// 	}, commentIndexIDKey, commentIndexStateAttrsObjIDObjTypeKey)
// 	return err
// }
