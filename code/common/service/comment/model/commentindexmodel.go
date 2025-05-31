package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
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
	}
)

// NewCommentIndexModel returns a model for the database table.
func NewCommentIndexModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentIndexModel {
	return &customCommentIndexModel{
		defaultCommentIndexModel: newCommentIndexModel(conn, c, opts...),
	}
}
