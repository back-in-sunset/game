package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentContent0Model = (*customCommentContent0Model)(nil)

var (
	cacheCommentContentObjIdCommentIdPrefix = "cache:commentContent:objId:commentId:"
)

type (
	// CommentContent0Model is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentContent0Model.
	CommentContent0Model interface {
		commentContent0Model
	}

	customCommentContent0Model struct {
		*defaultCommentContent0Model
		tableFn func(uint64) string
	}
)

func newCustomCommentContent0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultCommentContent0Model {
	return &defaultCommentContent0Model{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`comment_content_0`",
	}
}

// NewCommentContent0Model returns a model for the database table.
func NewCommentContent0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentContent0Model {
	return &customCommentContent0Model{
		defaultCommentContent0Model: newCustomCommentContent0Model(conn, c, opts...),
		tableFn: func(shardingId uint64) string {
			// Use the last 8 bits of the shardingId for determining the table suffix.
			const shardingBitmask = 0xFF // Adjust this bitmask if the sharding logic changes.
			return fmt.Sprintf("`comment_content_%d`", shardingId&shardingBitmask)
		},
	}
}

func (m *customCommentContent0Model) FindOne(ctx context.Context, commentId uint64) (*CommentContent0, error) {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return nil, fmt.Errorf("objId is required in context")
	}

	commentContentCommentIdKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIdCommentIdPrefix, objId, commentId)
	var resp CommentContent0
	err := m.QueryRowCtx(ctx, &resp, commentContentCommentIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `comment_id` = ? limit 1", commentContent0Rows, m.tableFn(commentId))
		return conn.QueryRowCtx(ctx, v, query, commentId)
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

func (m *customCommentContent0Model) Delete(ctx context.Context, commentId uint64) error {
	objId, ok := ctx.Value("objId").(uint64)
	if !ok || objId == 0 {
		return fmt.Errorf("objId is required in context")
	}

	commentContentCommentIdKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIdCommentIdPrefix, objId, commentId)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `comment_id` = ?", m.tableFn(commentId))
		return conn.ExecCtx(ctx, query, commentId)
	}, commentContentCommentIdKey)
	return err
}

func (m *customCommentContent0Model) Insert(ctx context.Context, data *CommentContent0) (sql.Result, error) {
	commentContentCommentIdKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIdCommentIdPrefix, data.ObjId, data.CommentId)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?)", m.tableFn(data.ObjId), commentContent0RowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.CommentId, data.AtMemberIds, data.Ip, data.Platform, data.Device, data.Massage, data.Meta)
	}, commentContentCommentIdKey)
	return ret, err
}

func (m *customCommentContent0Model) Update(ctx context.Context, data *CommentContent0) error {
	commentContentCommentIdKey := fmt.Sprintf("%s%v%v", cacheCommentContentObjIdCommentIdPrefix, data.ObjId, data.CommentId)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `comment_id` = ?", m.tableFn(data.ObjId), commentContent0RowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, data.AtMemberIds, data.Ip, data.Platform, data.Device, data.Massage, data.Meta, data.CommentId)
	}, commentContentCommentIdKey)
	return err
}
