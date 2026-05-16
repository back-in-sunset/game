//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"comment/rpc/comment"
	logicpkg "comment/rpc/internal/logic"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"game/server/core/testkit"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestGetCommentList_PinnedCommentsFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	objID := int64(400001)
	shard := objID & 0xFF

	ctx, dsn := testkit.StartMySQLContainer(t, "comment_pinned")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)

	// Create tables for this shard
	createTablesSQL := []string{
		fmt.Sprintf(`CREATE TABLE comment_subject_%d (
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
		fmt.Sprintf(`CREATE TABLE comment_index_%d (
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
		fmt.Sprintf(`CREATE TABLE comment_content_%d (
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

	for _, sql := range createTablesSQL {
		if _, err := db.Exec(sql); err != nil {
			t.Fatalf("create table failed: %v", err)
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
		CommentModel: commentModel,
		BizRedis:     rds,
		LikeEventBus: noopLikeEventBus{},
	}

	objType := int64(1)
	memberID := int64(6001)

	// Create 3 comments
	createLogic := logicpkg.NewAddCommentLogic(context.Background(), svcCtx)
	_, err = createLogic.AddComment(&comment.CommentRequest{
		ObjID:    objID,
		ObjType:  objType,
		MemberID: memberID,
		Message:  "First comment",
	})
	if err != nil {
		t.Fatalf("create comment1: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	comment2, err := createLogic.AddComment(&comment.CommentRequest{
		ObjID:    objID,
		ObjType:  objType,
		MemberID: memberID,
		Message:  "Pinned comment",
	})
	if err != nil {
		t.Fatalf("create comment2: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = createLogic.AddComment(&comment.CommentRequest{
		ObjID:    objID,
		ObjType:  objType,
		MemberID: memberID,
		Message:  "Last comment",
	})
	if err != nil {
		t.Fatalf("create comment3: %v", err)
	}

	// Pin comment2
	pinLogic := logicpkg.NewSetCommentAttrsLogic(context.Background(), svcCtx)
	_, err = pinLogic.SetCommentAttrs(&comment.SetCommentAttrsRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: comment2.CommentID,
		MemberID:  memberID,
	})
	if err != nil {
		t.Fatalf("pin comment2: %v", err)
	}

	// Wait for the pin to take effect and cache to update
	time.Sleep(100 * time.Millisecond)

	// Clear the Redis cache to test the database query fix
	timeSortKey := fmt.Sprintf("biz#commentids#objID:%d:objType:%d:rootID:0:sortType:0", objID, objType)
	rds.Del(timeSortKey)
	timeSortKeyCompat := fmt.Sprintf("biz#commentobj#sorttype#objID:%d:objType:%d:rootID:0:sortType:0", objID, objType)
	rds.Del(timeSortKeyCompat)

	// Get list sorted by time - pinned comment should be first
	listLogic := logicpkg.NewGetCommentListLogic(context.Background(), svcCtx)
	listResp, err := listLogic.GetCommentList(&comment.CommentListRequest{
		ObjID:    objID,
		ObjType:  objType,
		SortType: 0,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("get comment list: %v", err)
	}

	if len(listResp.Comments) < 3 {
		t.Fatalf("expected at least 3 comments, got %d", len(listResp.Comments))
	}

	// First comment should be pinned one
	if listResp.Comments[0].CommentID != comment2.CommentID {
		t.Errorf("expected pinned comment first, got comment id %d", listResp.Comments[0].CommentID)
	}
}