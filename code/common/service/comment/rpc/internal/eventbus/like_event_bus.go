package eventbus

import "context"

type LikeAction string

const (
	LikeActionLike   LikeAction = "like"
	LikeActionUnlike LikeAction = "unlike"
)

type LikeEvent struct {
	Action    LikeAction
	ObjID     int64
	ObjType   int64
	CommentID int64
	MemberID  int64
	Delta     int64
	Ts        int64
}

type LikeEventHandler func(ctx context.Context, event LikeEvent) error

type LikeEventBus interface {
	PublishLikeEvent(ctx context.Context, event LikeEvent) error
	ConsumeLikeEvents(ctx context.Context, consumerID string, handler LikeEventHandler)
	Close() error
}
