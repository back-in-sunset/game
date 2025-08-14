package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentContentModel = (*customCommentContentModel)(nil)

const (
	cacheCommentContentObjIDCommentIDPrefix = "cache:commentContent:objId:commentId:"
)

type (
	// CommentContentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentContentModel.
	CommentContentModel interface {
		// commentContentModel
		Insert(ctx context.Context, data *CommentContent) (sql.Result, error)
		FindOne(ctx context.Context, objeID uint64, commentID uint64) (*CommentContent, error)
		Update(ctx context.Context, data *CommentContent) error
		Delete(ctx context.Context, objeID uint64, commentID uint64) error
	}

	customCommentContentModel struct {
		// *defaultCommentContentModel
		sqlc.CachedConn
		tableFn func(uint64) string
	}
)

func newCustomCommentContentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *customCommentContentModel {
	return &customCommentContentModel{
		// defaultCommentContentModel: newCommentContentModel(conn, c, opts...),
		CachedConn: sqlc.NewConn(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_content_%d`", shardingId&shardingBitmask)
		},
	}
}

// NewCommentContentModel returns a model for the database table.
func NewCommentContentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentContentModel {
	return &customCommentContentModel{
		// defaultCommentContentModel: newCommentContentModel(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_content_%d`", shardingId&shardingBitmask)
		},
	}
}

func (m *customCommentContentModel) FindOne(ctx context.Context, objID uint64, commentID uint64) (*CommentContent, error) {
	commentContentCommentIDKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIDCommentIDPrefix, objID, commentID)
	var resp CommentContent
	err := m.QueryRowCtx(ctx, &resp, commentContentCommentIDKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `comment_id` = ? limit 1", commentContentRows, m.tableFn(objID))
		return conn.QueryRowCtx(ctx, v, query, commentID)
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

func (m *customCommentContentModel) Delete(ctx context.Context, objID uint64, commentID uint64) error {

	commentContentCommentIDKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIDCommentIDPrefix, objID, commentID)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `comment_id` = ?", m.tableFn(objID))
		return conn.ExecCtx(ctx, query, commentID)
	}, commentContentCommentIDKey)
	return err
}

func (m *customCommentContentModel) Insert(ctx context.Context, data *CommentContent) (sql.Result, error) {
	commentContentCommentIDKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIDCommentIDPrefix, data.ObjId, data.CommentId)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.tableFn(data.ObjId), commentContentRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.CommentId, data.ObjId, data.AtMemberIds, data.Ip, data.Platform, data.Device, data.Message, data.Meta)
	}, commentContentCommentIDKey)
	return ret, err
}

func (m *customCommentContentModel) Update(ctx context.Context, data *CommentContent) error {
	commentContentCommentIDKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIDCommentIDPrefix, data.ObjId, data.CommentId)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `comment_id` = ?", m.tableFn(data.ObjId), commentContentRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, data.AtMemberIds, data.Ip, data.Platform, data.Device, data.Message, data.Meta, data.CommentId)
	}, commentContentCommentIDKey)
	return err
}
