package eventbus

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	likeEventStream = "stream:comment:like:event"
	likeEventGroup  = "comment-like-persister"
)

type RedisLikeEventBus struct {
	rds  *redis.Redis
	node redis.ClosableNode
}

func NewRedisLikeEventBus(rds *redis.Redis) (*RedisLikeEventBus, error) {
	node, err := redis.CreateBlockingNode(rds)
	if err != nil {
		return nil, err
	}

	return &RedisLikeEventBus{
		rds:  rds,
		node: node,
	}, nil
}

func (b *RedisLikeEventBus) PublishLikeEvent(ctx context.Context, event LikeEvent) error {
	_, err := b.rds.XAddCtx(ctx, likeEventStream, false, "*", map[string]any{
		"action":     string(event.Action),
		"obj_id":     event.ObjID,
		"obj_type":   event.ObjType,
		"comment_id": event.CommentID,
		"member_id":  event.MemberID,
		"delta":      event.Delta,
		"ts":         event.Ts,
	})
	return err
}

func (b *RedisLikeEventBus) ConsumeLikeEvents(ctx context.Context, consumerID string, handler LikeEventHandler) {
	if err := b.ensureGroup(ctx); err != nil {
		logx.Errorf("ensure like event group error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := b.rds.XReadGroupCtx(ctx, b.node, likeEventGroup, consumerID, 64, 2*time.Second, false, likeEventStream, ">")
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if strings.Contains(err.Error(), "NOGROUP") {
				if ensureErr := b.ensureGroup(ctx); ensureErr != nil {
					logx.Errorf("recreate like event group error: %v", ensureErr)
				}
				continue
			}
			logx.Errorf("xreadgroup like event error: %v", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				event, parseErr := parseLikeEvent(msg.Values)
				if parseErr != nil {
					logx.Errorf("invalid like event fields: id=%s values=%v err=%v", msg.ID, msg.Values, parseErr)
					_, _ = b.rds.XAckCtx(ctx, likeEventStream, likeEventGroup, msg.ID)
					continue
				}

				if handleErr := handler(ctx, event); handleErr != nil {
					logx.Errorf("handle like event error: %v, msgID=%s", handleErr, msg.ID)
					continue
				}

				if _, ackErr := b.rds.XAckCtx(ctx, likeEventStream, likeEventGroup, msg.ID); ackErr != nil {
					logx.Errorf("xack like event error: %v, msgID=%s", ackErr, msg.ID)
				}
			}
		}
	}
}

func (b *RedisLikeEventBus) Close() error {
	if b.node != nil {
		b.node.Close()
	}
	return nil
}

func (b *RedisLikeEventBus) ensureGroup(ctx context.Context) error {
	_, err := b.rds.XGroupCreateMkStreamCtx(ctx, likeEventStream, likeEventGroup, "0")
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func parseLikeEvent(values map[string]any) (LikeEvent, error) {
	action, ok := values["action"]
	if !ok {
		return LikeEvent{}, fmt.Errorf("missing action")
	}

	objID, err := parseMessageInt64(values, "obj_id")
	if err != nil {
		return LikeEvent{}, err
	}
	objType, err := parseMessageInt64(values, "obj_type")
	if err != nil {
		return LikeEvent{}, err
	}
	commentID, err := parseMessageInt64(values, "comment_id")
	if err != nil {
		return LikeEvent{}, err
	}
	memberID, err := parseMessageInt64(values, "member_id")
	if err != nil {
		return LikeEvent{}, err
	}
	delta, err := parseMessageInt64(values, "delta")
	if err != nil {
		return LikeEvent{}, err
	}
	ts, err := parseMessageInt64(values, "ts")
	if err != nil {
		return LikeEvent{}, err
	}

	return LikeEvent{
		Action:    LikeAction(fmt.Sprintf("%v", action)),
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentID,
		MemberID:  memberID,
		Delta:     delta,
		Ts:        ts,
	}, nil
}

func parseMessageInt64(values map[string]any, key string) (int64, error) {
	v, ok := values[key]
	if !ok {
		return 0, fmt.Errorf("missing key %s", key)
	}

	switch x := v.(type) {
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	case float64:
		return int64(x), nil
	case string:
		return strconv.ParseInt(x, 10, 64)
	case []byte:
		return strconv.ParseInt(string(x), 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T for key %s", v, key)
	}
}
