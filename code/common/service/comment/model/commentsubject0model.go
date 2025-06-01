package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentSubject0Model = (*customCommentSubject0Model)(nil)

type (
	// CommentSubject0Model is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentSubject0Model.
	CommentSubject0Model interface {
		commentSubject0Model
	}

	customCommentSubject0Model struct {
		*defaultCommentSubject0Model
	}
)

// NewCommentSubject0Model returns a model for the database table.
func NewCommentSubject0Model(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentSubject0Model {
	return &customCommentSubject0Model{
		defaultCommentSubject0Model: newCommentSubject0Model(conn, c, opts...),
	}
}
