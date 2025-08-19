package types

const (
	// SortCreatedTime 按创建时间排序
	SortCreatedTime = iota
	// SortLikeCount 按点赞数排序
	SortLikeCount
)

const (
	// DefaultPageSize 默认分页大小
	DefaultPageSize = 20
	// DefaultLimit 默认限制
	DefaultLimit = 200
	// DefaultSortLikeCursor 默认点赞数游标
	DefaultSortLikeCursor = 1 << 30
)

const (
	// PrefixCommentObj
	PrefixCommentObj = "biz#comment#objid:%d"

	PrefixCommentIndex   = "comment:index:"
	PrefixCommentContent = "comment:content:"
)
