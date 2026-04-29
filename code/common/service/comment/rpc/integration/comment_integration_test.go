//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"comment/rpc/comment"
	"comment/rpc/internal/eventbus"
	logicpkg "comment/rpc/internal/logic"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"game/server/core/testkit"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type noopLikeEventBus struct{}

func (noopLikeEventBus) PublishLikeEvent(context.Context, eventbus.LikeEvent) error { return nil }
func (noopLikeEventBus) ConsumeLikeEvents(context.Context, string, eventbus.LikeEventHandler) {
}
func (noopLikeEventBus) Close() error { return nil }

type jsonCase struct {
	Name  string     `json:"name"`
	ObjID int64      `json:"obj_id"`
	Steps []jsonStep `json:"steps"`
}

type jsonStep struct {
	Op               string `json:"op"`
	Alias            string `json:"alias,omitempty"`
	ObjType          int64  `json:"obj_type,omitempty"`
	MemberID         int64  `json:"member_id,omitempty"`
	Message          string `json:"message,omitempty"`
	RootID           int64  `json:"root_id,omitempty"`
	RootAlias        string `json:"root_alias,omitempty"`
	ReplyID          int64  `json:"reply_id,omitempty"`
	ReplyAlias       string `json:"reply_alias,omitempty"`
	SortType         int64  `json:"sort_type,omitempty"`
	PageSize         int64  `json:"page_size,omitempty"`
	ExpectCount      int    `json:"expect_count,omitempty"`
	ExpectContains   string `json:"expect_contains,omitempty"`
	ExpectRootAlias  string `json:"expect_root_alias,omitempty"`
	ExpectReplyAlias string `json:"expect_reply_alias,omitempty"`
	ExpectState      int64  `json:"expect_state,omitempty"`
	CommentAlias     string `json:"comment_alias,omitempty"`
	ExpectTopAlias   string `json:"expect_top_alias,omitempty"`
	LikeMemberID     int64  `json:"like_member_id,omitempty"`
	ExpectError      string `json:"expect_error,omitempty"`
}

const nestedCommentJSONCases = `
[
  {
    "name": "publish_root_and_reply_then_query",
    "obj_id": 131073,
    "steps": [
      {
        "op": "add",
        "alias": "root",
        "obj_type": 1,
        "member_id": 2001,
        "message": "root comment from json"
      },
      {
        "op": "add",
        "alias": "reply",
        "obj_type": 1,
        "member_id": 2002,
        "message": "reply comment from json",
        "root_alias": "root",
        "reply_alias": "root"
      },
      {
        "op": "list",
        "obj_type": 1,
        "sort_type": 0,
        "page_size": 10,
        "expect_count": 1,
        "expect_contains": "root comment from json"
      },
      {
        "op": "list",
        "obj_type": 1,
        "root_alias": "root",
        "sort_type": 0,
        "page_size": 10,
        "expect_count": 1,
        "expect_contains": "reply comment from json",
        "expect_root_alias": "root",
        "expect_reply_alias": "root"
      }
    ]
  },
  {
    "name": "delete_root_then_reply_still_queryable",
    "obj_id": 131329,
    "steps": [
      {
        "op": "add",
        "alias": "root",
        "obj_type": 1,
        "member_id": 2101,
        "message": "root to delete"
      },
      {
        "op": "add",
        "alias": "reply",
        "obj_type": 1,
        "member_id": 2102,
        "message": "reply remains visible",
        "root_alias": "root",
        "reply_alias": "root"
      },
      {
        "op": "delete",
        "obj_type": 1,
        "member_id": 2101,
        "comment_alias": "root",
        "expect_state": 1
      },
      {
        "op": "list",
        "obj_type": 1,
        "root_alias": "root",
        "sort_type": 0,
        "page_size": 10,
        "expect_count": 1,
        "expect_contains": "reply remains visible",
        "expect_root_alias": "root",
        "expect_reply_alias": "root"
      }
    ]
  },
  {
    "name": "like_then_sort_by_hot",
    "obj_id": 131585,
    "steps": [
      {
        "op": "add",
        "alias": "c1",
        "obj_type": 1,
        "member_id": 2201,
        "message": "comment low heat"
      },
      {
        "op": "add",
        "alias": "c2",
        "obj_type": 1,
        "member_id": 2202,
        "message": "comment high heat"
      },
      {
        "op": "like",
        "obj_type": 1,
        "comment_alias": "c2",
        "like_member_id": 3001
      },
      {
        "op": "like",
        "obj_type": 1,
        "comment_alias": "c2",
        "like_member_id": 3002
      },
      {
        "op": "list",
        "obj_type": 1,
        "sort_type": 1,
        "page_size": 10,
        "expect_count": 2,
        "expect_top_alias": "c2"
      }
    ]
  },
  {
    "name": "reject_reply_without_root",
    "obj_id": 131841,
    "steps": [
      {
        "op": "add",
        "alias": "root",
        "obj_type": 1,
        "member_id": 2301,
        "message": "root for invalid reply"
      },
      {
        "op": "add",
        "obj_type": 1,
        "member_id": 2302,
        "message": "invalid reply without root",
        "reply_alias": "root",
        "expect_error": "invalid reply relation"
      }
    ]
  }
]
`

func TestNestedCommentFlow_JSONTableDriven_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	var cases []jsonCase
	if err := json.Unmarshal([]byte(nestedCommentJSONCases), &cases); err != nil {
		t.Fatalf("unmarshal json cases: %v", err)
	}
	if len(cases) == 0 {
		t.Fatalf("empty json cases")
	}

	ctx, dsn := testkit.StartMySQLContainer(t, "comment")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)
	mustCreateCommentTables(t, db, cases)

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

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			idsByAlias := map[string]int64{}

			for i, step := range tc.Steps {
				switch step.Op {
				case "add":
					rootID := step.RootID
					replyID := step.ReplyID
					if step.RootAlias != "" {
						rootID = idsByAlias[step.RootAlias]
					}
					if step.ReplyAlias != "" {
						replyID = idsByAlias[step.ReplyAlias]
					}

					addLogic := logicpkg.NewAddCommentLogic(context.Background(), svcCtx)
					resp, addErr := addLogic.AddComment(&comment.CommentRequest{
						ObjID:    tc.ObjID,
						ObjType:  step.ObjType,
						MemberID: step.MemberID,
						Message:  step.Message,
						RootID:   rootID,
						ReplyID:  replyID,
					})
					if step.ExpectError != "" {
						if addErr == nil {
							t.Fatalf("step[%d] add error=nil want contains=%q", i, step.ExpectError)
						}
						if !strings.Contains(addErr.Error(), step.ExpectError) {
							t.Fatalf("step[%d] add error=%v want contains=%q", i, addErr, step.ExpectError)
						}
						continue
					}
					if addErr != nil {
						t.Fatalf("step[%d] add error=%v", i, addErr)
					}
					if resp.CommentID <= 0 {
						t.Fatalf("step[%d] invalid comment id=%d", i, resp.CommentID)
					}
					if step.Alias != "" {
						idsByAlias[step.Alias] = resp.CommentID
					}

				case "list":
					rootID := step.RootID
					if step.RootAlias != "" {
						rootID = idsByAlias[step.RootAlias]
					}

					listLogic := logicpkg.NewGetCommentListLogic(context.Background(), svcCtx)
					resp, listErr := listLogic.GetCommentList(&comment.CommentListRequest{
						ObjID:    tc.ObjID,
						ObjType:  step.ObjType,
						RootID:   rootID,
						SortType: step.SortType,
						PageSize: step.PageSize,
					})
					if listErr != nil {
						t.Fatalf("step[%d] list error=%v", i, listErr)
					}
					if len(resp.Comments) != step.ExpectCount {
						t.Fatalf("step[%d] comments count=%d want=%d", i, len(resp.Comments), step.ExpectCount)
					}

					joined := make([]string, 0, len(resp.Comments))
					for _, c := range resp.Comments {
						joined = append(joined, c.Message)
					}
					if step.ExpectContains != "" && !strings.Contains(strings.Join(joined, "|"), step.ExpectContains) {
						t.Fatalf("step[%d] comments=%q should contain=%q", i, strings.Join(joined, "|"), step.ExpectContains)
					}

					if step.ExpectRootAlias != "" || step.ExpectReplyAlias != "" {
						if len(resp.Comments) == 0 {
							t.Fatalf("step[%d] empty comments", i)
						}
						comment0 := resp.Comments[0]
						if step.ExpectRootAlias != "" && comment0.RootID != idsByAlias[step.ExpectRootAlias] {
							t.Fatalf("step[%d] root_id=%d want=%d", i, comment0.RootID, idsByAlias[step.ExpectRootAlias])
						}
						if step.ExpectReplyAlias != "" && comment0.ReplyID != idsByAlias[step.ExpectReplyAlias] {
							t.Fatalf("step[%d] reply_id=%d want=%d", i, comment0.ReplyID, idsByAlias[step.ExpectReplyAlias])
						}
					}
					if step.ExpectTopAlias != "" {
						if len(resp.Comments) == 0 {
							t.Fatalf("step[%d] empty comments", i)
						}
						if resp.Comments[0].CommentID != idsByAlias[step.ExpectTopAlias] {
							t.Fatalf("step[%d] top comment=%d want=%d", i, resp.Comments[0].CommentID, idsByAlias[step.ExpectTopAlias])
						}
					}

				case "delete":
					commentID := idsByAlias[step.CommentAlias]
					delLogic := logicpkg.NewDeleteCommentLogic(context.Background(), svcCtx)
					resp, delErr := delLogic.DeleteComment(&comment.CommentRequest{
						ObjID:     tc.ObjID,
						ObjType:   step.ObjType,
						CommentID: commentID,
						MemberID:  step.MemberID,
					})
					if delErr != nil {
						t.Fatalf("step[%d] delete error=%v", i, delErr)
					}
					if resp.State != step.ExpectState {
						t.Fatalf("step[%d] state=%d want=%d", i, resp.State, step.ExpectState)
					}

				case "like":
					commentID := idsByAlias[step.CommentAlias]
					likeLogic := logicpkg.NewLikeCommentLogic(context.Background(), svcCtx)
					_, likeErr := likeLogic.LikeComment(&comment.LikeCommentRequest{
						ObjID:     tc.ObjID,
						ObjType:   step.ObjType,
						CommentID: commentID,
						MemberID:  step.LikeMemberID,
					})
					if likeErr != nil {
						t.Fatalf("step[%d] like error=%v", i, likeErr)
					}

				default:
					t.Fatalf("step[%d] unsupported op=%s", i, step.Op)
				}
			}
		})
	}
}

func mustCreateCommentTables(t *testing.T, db *sql.DB, cases []jsonCase) {
	t.Helper()

	created := map[int64]struct{}{}
	for _, tc := range cases {
		shard := tc.ObjID & 0xFF
		if _, ok := created[shard]; ok {
			continue
		}
		created[shard] = struct{}{}

		stmts := []string{
			fmt.Sprintf(`
CREATE TABLE comment_subject_%d (
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
CREATE TABLE comment_index_%d (
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
CREATE TABLE comment_content_%d (
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
}
