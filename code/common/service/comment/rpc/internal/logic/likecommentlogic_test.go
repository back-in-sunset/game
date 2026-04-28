package logic

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/eventbus"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type likeCommentStubModel struct {
	findResp *model.Comment
	findErr  error
}

func (m *likeCommentStubModel) AddComment(context.Context, *model.CommentSubject, *model.CommentIndex, *model.CommentContent) (*model.CommentSchema, error) {
	return nil, nil
}

func (m *likeCommentStubModel) DeleteComment(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *likeCommentStubModel) CommentListByObjID(context.Context, int64, int64, int64, int64, string, int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *likeCommentStubModel) FindOneByObjID(context.Context, int64, int64) (*model.Comment, error) {
	return m.findResp, m.findErr
}

func (m *likeCommentStubModel) CacheCommentsByIDs(context.Context, int64, []int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *likeCommentStubModel) AdjustCommentLikeCount(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (m *likeCommentStubModel) SetCommentState(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *likeCommentStubModel) SetCommentAttrs(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

type fakeLikeEventBus struct {
	mu     sync.Mutex
	events []eventbus.LikeEvent
}

func (b *fakeLikeEventBus) PublishLikeEvent(_ context.Context, event eventbus.LikeEvent) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, event)
	return nil
}

func (b *fakeLikeEventBus) ConsumeLikeEvents(context.Context, string, eventbus.LikeEventHandler) {}

func (b *fakeLikeEventBus) Close() error { return nil }

func (b *fakeLikeEventBus) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.events)
}

func TestLikeCommentLogic_ConcurrentIdempotent(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable at 127.0.0.1:6379")
	}

	rds, err := redis.NewRedis(redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
	})
	if err != nil {
		t.Fatalf("create redis client: %v", err)
	}

	objID := time.Now().UnixNano()
	objType := int64(1)
	commentID := int64(9001)
	memberID := int64(2001)
	rootID := int64(0)

	likedUsersKey := likeKeyLikedUsers(objID, commentID)
	likeKey := likeKeyByLikeScore(objID, objType, rootID)
	likeCompatKey := likeKeyByLikeScoreCompat(objID, objType, rootID)
	if _, err = rds.EvalCtx(context.Background(), `return redis.call("DEL", KEYS[1], KEYS[2], KEYS[3])`, []string{likedUsersKey, likeKey, likeCompatKey}); err != nil {
		t.Fatalf("cleanup redis keys: %v", err)
	}
	defer func() {
		_, _ = rds.EvalCtx(context.Background(), `return redis.call("DEL", KEYS[1], KEYS[2], KEYS[3])`, []string{likedUsersKey, likeKey, likeCompatKey})
	}()

	bus := &fakeLikeEventBus{}
	l := NewLikeCommentLogic(context.Background(), &svc.ServiceContext{
		CommentModel: &likeCommentStubModel{
			findResp: &model.Comment{
				ID:        commentID,
				ObjID:     objID,
				ObjType:   objType,
				RootID:    rootID,
				Attrs:     0,
				CreatedAt: time.Now(),
			},
		},
		BizRedis:     rds,
		LikeEventBus: bus,
	})

	const workers = 30
	errCh := make(chan error, workers)
	msgCh := make(chan string, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, callErr := l.LikeComment(&comment.LikeCommentRequest{
				ObjID:     objID,
				ObjType:   objType,
				CommentID: commentID,
				MemberID:  memberID,
			})
			if callErr != nil {
				errCh <- callErr
				return
			}
			msgCh <- resp.Message
		}()
	}
	wg.Wait()
	close(errCh)
	close(msgCh)

	for callErr := range errCh {
		t.Fatalf("LikeComment() error = %v", callErr)
	}

	okCount := 0
	alreadyLikedCount := 0
	for msg := range msgCh {
		if msg == "ok" {
			okCount++
		}
		if msg == "already liked" {
			alreadyLikedCount++
		}
	}
	if okCount != 1 {
		t.Fatalf("ok message count=%d want=1", okCount)
	}
	if alreadyLikedCount != workers-1 {
		t.Fatalf("already liked count=%d want=%d", alreadyLikedCount, workers-1)
	}
	if bus.Count() != 1 {
		t.Fatalf("publish like event count=%d want=1", bus.Count())
	}

	res, err := rds.EvalCtx(context.Background(), `
local likedUsers = redis.call("SCARD", KEYS[1])
local score = redis.call("ZSCORE", KEYS[2], ARGV[1])
local compatScore = redis.call("ZSCORE", KEYS[3], ARGV[1])
return {likedUsers, score, compatScore}
`, []string{likedUsersKey, likeKey, likeCompatKey}, "9001")
	if err != nil {
		t.Fatalf("query redis state: %v", err)
	}

	items, ok := res.([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("unexpected eval result: %T %v", res, res)
	}
	likedUsers, err := castToInt64(items[0])
	if err != nil {
		t.Fatalf("parse liked users: %v", err)
	}
	score, err := castToInt64(items[1])
	if err != nil {
		t.Fatalf("parse score: %v", err)
	}
	compatScore, err := castToInt64(items[2])
	if err != nil {
		t.Fatalf("parse compat score: %v", err)
	}

	if likedUsers != 1 {
		t.Fatalf("liked users=%d want=1", likedUsers)
	}
	if score != 1 {
		t.Fatalf("score=%d want=1", score)
	}
	if compatScore != 1 {
		t.Fatalf("compatScore=%d want=1", compatScore)
	}
}

func portReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
