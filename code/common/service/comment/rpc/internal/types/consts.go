package types

const (
	SortLikeCount = iota
	SortPublishTime
)

const (
	DefaultPageSize = 20
	DefaultLimit    = 200

	DefaultSortLikeCursor = 1 << 30
)
