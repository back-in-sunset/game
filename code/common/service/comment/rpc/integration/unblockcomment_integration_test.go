//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"comment/rpc/comment"
	logicpkg "comment/rpc/internal/logic"
	"comment/rpc/internal/notify"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"game/server/core/testkit"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestUnBlockComment_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	ctx, dsn := testkit.StartMySQLContainer(t, "comment_unblock")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)

	testObjIDs := []int64{200001, 200002, 200003, 200004, 200005, 200006}
	for _, objID := range testObjIDs {
		shard := objID & 0xFF
		stmts := []string{
			fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS comment_subject_%d (
  id bigint NOT NULL AUTO_INCREMENT,
  obj_id bigint NOT NULL DEFAULT 0,
  obj_type tinyint(3) NOT NULL DEFAULT 0,
  member_id bigint NOT NULL DEFAULT 0,
  count int(11) NOT NULL DEFAULT 0,
  root_count int(11) NOT NULL DEFAULT 0,
  all_count int(11) NOT NULL DEFAULT 0,
  state tinyint(3) NOT NULL DEFAULT 0,
  attrs int(11) NOT NULL DEFAULT 0,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY idx_obj_type_unique (state, obj_id, obj_type),
  KEY idx_member_unique (state, member_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, shard),
			fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS comment_index_%d (
  id bigint NOT NULL AUTO_INCREMENT,
  obj_id bigint NOT NULL DEFAULT 0,
  obj_type tinyint(3) NOT NULL DEFAULT 0,
  member_id bigint NOT NULL DEFAULT 0,
  root_id bigint NOT NULL DEFAULT 0,
  reply_id bigint NOT NULL DEFAULT 0,
  floor bigint NOT NULL DEFAULT 0,
  count int(11) NOT NULL DEFAULT 0,
  root_count int(11) NOT NULL DEFAULT 0,
  like_count int(11) NOT NULL DEFAULT 0,
  hate_count int(11) NOT NULL DEFAULT 0,
  state tinyint(3) NOT NULL DEFAULT 0,
  attrs int(11) NOT NULL DEFAULT 0,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_state_attrs_obj_type_unique (state, attrs, obj_id, obj_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, shard),
			fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS comment_content_%d (
  comment_id bigint NOT NULL,
  obj_id bigint NOT NULL DEFAULT 0,
  at_member_ids text NOT NULL,
  ip varchar(255) NOT NULL DEFAULT '',
  platform tinyint(3) NOT NULL DEFAULT 0,
  device varchar(255) NOT NULL DEFAULT '',
  message text NOT NULL,
  meta text NOT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (comment_id),
  KEY idx_comment_obj_unique (comment_id, obj_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, shard),
		}
		for _, stmt := range stmts {
			if _, err := db.Exec(stmt); err != nil {
				t.Fatalf("create table failed: %v, sql=%s", err, stmt)
			}
		}
	}

	redisAddr := testkit.StartMiniRedis(t)
	rds, err := redis.NewRedis(redis.RedisConf{
		Host: redisAddr,
		Type: "node",
	})
	if err != nil {
		t.Fatalf("new redis: %v", err)
	}

	commentModel := model.NewCommentModel(sqlx.NewMysql(dsn), cache.CacheConf{
		{
			Weight: 100,
			RedisConf: redis.RedisConf{
				Host: redisAddr,
				Type: "node",
			},
		},
	})
	svcCtx := &svc.ServiceContext{
		CommentModel:    commentModel,
		BizRedis:        rds,
		LikeEventBus:    noopLikeEventBus{},
		CommentNotifier: notify.NoopCommentNotifier{},
	}

	t.Run("unblock_comment_successfully", func(t *testing.T) {
		objID := int64(200001)
		objType := int64(1)
		memberID := int64(4001)

		addLogic := logicpkg.NewAddCommentLogic(context.Background(), svcCtx)
		addResp, err := addLogic.AddComment(&comment.CommentRequest{
			ObjID:    objID,
			ObjType:  objType,
			MemberID: memberID,
			Message:  "test comment for unblocking",
		})
		if err != nil {
			t.Fatalf("AddComment() error = %v", err)
		}
		if addResp.CommentID <= 0 {
			t.Fatalf("invalid comment id=%d", addResp.CommentID)
		}

		blockLogic := logicpkg.NewBlockCommentLogic(context.Background(), svcCtx)
		_, err = blockLogic.BlockComment(&comment.BlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})
		if err != nil {
			t.Fatalf("BlockComment() error = %v", err)
		}

		blockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if blockedComment.State != 1 {
			t.Fatalf("expected state=1 after block, got %d", blockedComment.State)
		}

		unblockLogic := logicpkg.NewUnBlockCommentLogic(context.Background(), svcCtx)
		unblockResp, err := unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})

		if err != nil {
			t.Fatalf("UnBlockComment() error = %v", err)
		}

		if !unblockResp.Success {
			t.Errorf("expected success=true, got %v", unblockResp.Success)
		}

		if unblockResp.Message != "ok" {
			t.Errorf("expected message='ok', got %v", unblockResp.Message)
		}

		unblockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}

		if unblockedComment.State != 0 {
			t.Errorf("expected state=0 (unblocked), got %d", unblockedComment.State)
		}
	})

	t.Run("unblock_already_unblocked_comment", func(t *testing.T) {
		objID := int64(200002)
		objType := int64(1)
		memberID := int64(4002)

		addLogic := logicpkg.NewAddCommentLogic(context.Background(), svcCtx)
		addResp, err := addLogic.AddComment(&comment.CommentRequest{
			ObjID:    objID,
			ObjType:  objType,
			MemberID: memberID,
			Message:  "test comment for double unblocking",
		})
		if err != nil {
			t.Fatalf("AddComment() error = %v", err)
		}

		unblockLogic := logicpkg.NewUnBlockCommentLogic(context.Background(), svcCtx)
		_, err = unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})
		if err != nil {
			t.Fatalf("First UnBlockComment() error = %v", err)
		}

		unblockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if unblockedComment.State != 0 {
			t.Fatalf("expected state=0 after first unblock, got %d", unblockedComment.State)
		}

		secondUnblockResp, err := unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})

		if err != nil {
			t.Fatalf("Second UnBlockComment() error = %v", err)
		}

		if !secondUnblockResp.Success {
			t.Errorf("expected success=true for second unblock, got %v", secondUnblockResp.Success)
		}

		stillUnblockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if stillUnblockedComment.State != 0 {
			t.Errorf("expected state=0 (still unblocked), got %d", stillUnblockedComment.State)
		}
	})

	t.Run("unblock_nonexistent_comment", func(t *testing.T) {
		objID := int64(200003)
		objType := int64(1)
		memberID := int64(4003)

		unblockLogic := logicpkg.NewUnBlockCommentLogic(context.Background(), svcCtx)
		_, err := unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: 99999,
			MemberID:  memberID,
		})

		if err == nil {
			t.Fatalf("expected error for non-existent comment, got nil")
		}
	})

	t.Run("unblock_multiple_comments", func(t *testing.T) {
		objID := int64(200004)
		objType := int64(1)
		memberID := int64(4004)

		addLogic := logicpkg.NewAddCommentLogic(context.Background(), svcCtx)

		comments := make([]int64, 3)
		for i := 0; i < 3; i++ {
			resp, err := addLogic.AddComment(&comment.CommentRequest{
				ObjID:    objID,
				ObjType:  objType,
				MemberID: memberID,
				Message:  "test comment for multiple operations",
			})
			if err != nil {
				t.Fatalf("AddComment() error = %v", err)
			}
			comments[i] = resp.CommentID
		}

		blockLogic := logicpkg.NewBlockCommentLogic(context.Background(), svcCtx)
		for _, commentID := range comments {
			_, err := blockLogic.BlockComment(&comment.BlockCommentRequest{
				ObjID:     objID,
				ObjType:   objType,
				CommentID: commentID,
				MemberID:  memberID,
			})
			if err != nil {
				t.Fatalf("BlockComment() error = %v for comment %d", err, commentID)
			}
		}

		unblockLogic := logicpkg.NewUnBlockCommentLogic(context.Background(), svcCtx)
		for _, commentID := range comments {
			resp, err := unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
				ObjID:     objID,
				ObjType:   objType,
				CommentID: commentID,
				MemberID:  memberID,
			})
			if err != nil {
				t.Fatalf("UnBlockComment() error = %v for comment %d", err, commentID)
			}
			if !resp.Success {
				t.Errorf("expected success=true, got %v", resp.Success)
			}
		}

		for _, commentID := range comments {
			commentData, err := commentModel.FindOneByObjID(context.Background(), objID, commentID)
			if err != nil {
				t.Fatalf("FindOneByObjID() error = %v for comment %d", err, commentID)
			}
			if commentData.State != 0 {
				t.Errorf("expected state=0 (unblocked), got %d for comment %d", commentData.State, commentID)
			}
		}
	})

	t.Run("unblock_comment_validation", func(t *testing.T) {
		objID := int64(200005)
		objType := int64(1)

		unblockLogic := logicpkg.NewUnBlockCommentLogic(context.Background(), svcCtx)

		_, err := unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjType:   objType,
			CommentID: 1,
			MemberID:  1,
		})
		if err == nil {
			t.Fatalf("expected error for missing ObjID, got nil")
		}

		_, err = unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:    objID,
			ObjType:  objType,
			MemberID: 1,
		})
		if err == nil {
			t.Fatalf("expected error for missing CommentID, got nil")
		}
	})

	t.Run("block_unblock_cycle", func(t *testing.T) {
		objID := int64(200006)
		objType := int64(1)
		memberID := int64(4005)

		addLogic := logicpkg.NewAddCommentLogic(context.Background(), svcCtx)
		addResp, err := addLogic.AddComment(&comment.CommentRequest{
			ObjID:    objID,
			ObjType:  objType,
			MemberID: memberID,
			Message:  "test comment for block/unblock cycle",
		})
		if err != nil {
			t.Fatalf("AddComment() error = %v", err)
		}

		blockLogic := logicpkg.NewBlockCommentLogic(context.Background(), svcCtx)
		unblockLogic := logicpkg.NewUnBlockCommentLogic(context.Background(), svcCtx)

		_, err = blockLogic.BlockComment(&comment.BlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})
		if err != nil {
			t.Fatalf("BlockComment() error = %v", err)
		}

		blockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if blockedComment.State != 1 {
			t.Fatalf("expected state=1 after block, got %d", blockedComment.State)
		}

		_, err = unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})
		if err != nil {
			t.Fatalf("UnBlockComment() error = %v", err)
		}

		unblockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if unblockedComment.State != 0 {
			t.Fatalf("expected state=0 after unblock, got %d", unblockedComment.State)
		}

		_, err = blockLogic.BlockComment(&comment.BlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})
		if err != nil {
			t.Fatalf("Second BlockComment() error = %v", err)
		}

		reblockedComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if reblockedComment.State != 1 {
			t.Fatalf("expected state=1 after reblock, got %d", reblockedComment.State)
		}

		_, err = unblockLogic.UnBlockComment(&comment.UnBlockCommentRequest{
			ObjID:     objID,
			ObjType:   objType,
			CommentID: addResp.CommentID,
			MemberID:  memberID,
		})
		if err != nil {
			t.Fatalf("Second UnBlockComment() error = %v", err)
		}

		finalComment, err := commentModel.FindOneByObjID(context.Background(), objID, addResp.CommentID)
		if err != nil {
			t.Fatalf("FindOneByObjID() error = %v", err)
		}
		if finalComment.State != 0 {
			t.Fatalf("expected state=0 after final unblock, got %d", finalComment.State)
		}
	})
}