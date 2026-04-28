package logic

import (
	"context"
	"sync"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestUnLikeCommentLogic_ConcurrentIdempotent(t *testing.T) {
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
	commentID := int64(9002)
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
	modelStub := &likeCommentStubModel{
		findResp: &model.Comment{
			ID:        commentID,
			ObjID:     objID,
			ObjType:   objType,
			RootID:    rootID,
			Attrs:     0,
			CreatedAt: time.Now(),
		},
	}
	serviceCtx := &svc.ServiceContext{
		CommentModel: modelStub,
		BizRedis:     rds,
		LikeEventBus: bus,
	}

	likeLogic := NewLikeCommentLogic(context.Background(), serviceCtx)
	_, err = likeLogic.LikeComment(&comment.LikeCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentID,
		MemberID:  memberID,
	})
	if err != nil {
		t.Fatalf("prepare like failed: %v", err)
	}
	if bus.Count() != 1 {
		t.Fatalf("prepare like event count=%d want=1", bus.Count())
	}

	unlikeLogic := NewUnLikeCommentLogic(context.Background(), serviceCtx)
	const workers = 20
	errCh := make(chan error, workers)
	msgCh := make(chan string, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, callErr := unlikeLogic.UnLikeComment(&comment.UnLikeCommentRequest{
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
		t.Fatalf("UnLikeComment() error = %v", callErr)
	}

	okCount := 0
	alreadyUnlikedCount := 0
	for msg := range msgCh {
		if msg == "ok" {
			okCount++
		}
		if msg == "already unliked" {
			alreadyUnlikedCount++
		}
	}
	if okCount != 1 {
		t.Fatalf("ok message count=%d want=1", okCount)
	}
	if alreadyUnlikedCount != workers-1 {
		t.Fatalf("already unliked count=%d want=%d", alreadyUnlikedCount, workers-1)
	}
	if bus.Count() != 2 {
		t.Fatalf("publish event count=%d want=2", bus.Count())
	}

	res, err := rds.EvalCtx(context.Background(), `
local likedUsers = redis.call("SCARD", KEYS[1])
local score = redis.call("ZSCORE", KEYS[2], ARGV[1])
local compatScore = redis.call("ZSCORE", KEYS[3], ARGV[1])
return {likedUsers, score, compatScore}
`, []string{likedUsersKey, likeKey, likeCompatKey}, "9002")
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

	if likedUsers != 0 {
		t.Fatalf("liked users=%d want=0", likedUsers)
	}
	if score != 0 {
		t.Fatalf("score=%d want=0", score)
	}
	if compatScore != 0 {
		t.Fatalf("compatScore=%d want=0", compatScore)
	}
}
