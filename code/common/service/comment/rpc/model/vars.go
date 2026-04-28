package model

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const ()

// ErrNotFound 评论不存在
var ErrNotFound = sqlx.ErrNotFound

// ErrPermissionDenied 表示无权限操作评论
var ErrPermissionDenied = errors.New("permission denied")
