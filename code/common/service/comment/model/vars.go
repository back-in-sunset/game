package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

const (
	// PrefixCommentObjSortType 评论前缀
	PrefixCommentObjSortType                      = "store#comment#objid:sorttype:%d:%d"
	cacheCommentIndexStateAttrsObjIDObjTypePrefix = ""
)

var (
	objIDStruct struct{}
)

// ErrNotFound 评论不存在
var ErrNotFound = sqlx.ErrNotFound
