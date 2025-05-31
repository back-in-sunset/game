package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentContentModel = (*customCommentContentModel)(nil)

type (
	// CommentContentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentContentModel.
	CommentContentModel interface {
		commentContentModel
	}

	customCommentContentModel struct {
		*defaultCommentContentModel
	}
)

// NewCommentContentModel returns a model for the database table.
func NewCommentContentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentContentModel {
	return &customCommentContentModel{
		defaultCommentContentModel: newCommentContentModel(conn, c, opts...),
	}
}
