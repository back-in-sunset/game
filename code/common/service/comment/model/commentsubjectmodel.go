package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentSubjectModel = (*customCommentSubjectModel)(nil)

type (
	// CommentSubjectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentSubjectModel.
	CommentSubjectModel interface {
		commentSubjectModel
	}

	customCommentSubjectModel struct {
		*defaultCommentSubjectModel
	}
)

// NewCommentSubjectModel returns a model for the database table.
func NewCommentSubjectModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentSubjectModel {
	return &customCommentSubjectModel{
		defaultCommentSubjectModel: newCommentSubjectModel(conn, c, opts...),
	}
}
