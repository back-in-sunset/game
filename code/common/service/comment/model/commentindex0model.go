package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
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
	}
)

// NewCommentIndex0Model returns a model for the database table.
func NewCommentIndex0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentIndex0Model {
	return &customCommentIndex0Model{
		defaultCommentIndex0Model: newCommentIndex0Model(conn, c, opts...),
	}
}
