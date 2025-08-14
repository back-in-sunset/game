package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var (
	objIDStruct struct{}
)

// ErrNotFound 评论不存在
var ErrNotFound = sqlx.ErrNotFound
