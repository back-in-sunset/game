//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"game/server/core/testkit"
	"history/model"
	"history/rpc/historyclient"
	logicpkg "history/rpc/internal/logic"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type jsonCase struct {
	Name  string     `json:"name"`
	Steps []jsonStep `json:"steps"`
}

type jsonStep struct {
	Op          string `json:"op"`
	Alias       string `json:"alias,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	MediaType   int64  `json:"media_type,omitempty"`
	MediaID     int64  `json:"media_id,omitempty"`
	Title       string `json:"title,omitempty"`
	ProgressMs  int64  `json:"progress_ms,omitempty"`
	DurationMs  int64  `json:"duration_ms,omitempty"`
	Finished    bool   `json:"finished,omitempty"`
	PageSize    int64  `json:"page_size,omitempty"`
	ExpectCount int    `json:"expect_count,omitempty"`
	ExpectTitle string `json:"expect_title,omitempty"`
	ExpectDone  bool   `json:"expect_done,omitempty"`
}

const historyJSONCases = `
[
  {
    "name": "record_update_list_delete_clear",
    "steps": [
      {"op":"record","alias":"post","user_id":1001,"media_type":1,"media_id":2001,"title":"post first"},
      {"op":"record","alias":"video","user_id":1001,"media_type":2,"media_id":3001,"title":"video first","progress_ms":120000,"duration_ms":600000},
      {"op":"record","alias":"video","user_id":1001,"media_type":2,"media_id":3001,"title":"video updated","progress_ms":600000,"duration_ms":600000,"finished":true},
      {"op":"list","user_id":1001,"page_size":10,"expect_count":2,"expect_title":"video updated","expect_done":true},
      {"op":"list","user_id":1001,"media_type":2,"page_size":10,"expect_count":1,"expect_title":"video updated","expect_done":true},
      {"op":"delete_item","user_id":1001,"media_type":2,"media_id":3001},
      {"op":"list","user_id":1001,"media_type":2,"page_size":10,"expect_count":0},
      {"op":"clear_type","user_id":1001,"media_type":1},
      {"op":"list","user_id":1001,"page_size":10,"expect_count":0},
      {"op":"record","user_id":1001,"media_type":1,"media_id":2002,"title":"post second"},
      {"op":"clear_all","user_id":1001},
      {"op":"list","user_id":1001,"page_size":10,"expect_count":0}
    ]
  }
]
`

func TestHistoryFlow_JSONTableDriven_Integration(t *testing.T) {
	var cases []jsonCase
	if err := json.Unmarshal([]byte(historyJSONCases), &cases); err != nil {
		t.Fatalf("unmarshal json cases: %v", err)
	}

	ctx, dsn := testkit.StartMySQLContainer(t, "history")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)
	mustCreateHistoryTable(t, db)

	svcCtx := &svc.ServiceContext{
		HistoryModel: model.NewHistoryModel(sqlx.NewMysql(dsn)),
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for i, step := range tc.Steps {
				switch step.Op {
				case "record":
					l := logicpkg.NewRecordHistoryLogic(context.Background(), svcCtx)
					resp, err := l.RecordHistory(&historyclient.RecordHistoryRequest{
						UserID:     step.UserID,
						MediaType:  step.MediaType,
						MediaID:    step.MediaID,
						Title:      step.Title,
						ProgressMs: step.ProgressMs,
						DurationMs: step.DurationMs,
						Finished:   step.Finished,
					})
					if err != nil {
						t.Fatalf("step[%d] record error: %v", i, err)
					}
					if resp.Record == nil || resp.Record.ID <= 0 {
						t.Fatalf("step[%d] invalid record response: %+v", i, resp)
					}
				case "list":
					l := logicpkg.NewListHistoryLogic(context.Background(), svcCtx)
					resp, err := l.ListHistory(&historyclient.ListHistoryRequest{
						UserID:    step.UserID,
						MediaType: step.MediaType,
						PageSize:  step.PageSize,
					})
					if err != nil {
						t.Fatalf("step[%d] list error: %v", i, err)
					}
					if len(resp.List) != step.ExpectCount {
						t.Fatalf("step[%d] list count=%d want=%d", i, len(resp.List), step.ExpectCount)
					}
					if step.ExpectTitle != "" {
						joined := make([]string, 0, len(resp.List))
						for _, item := range resp.List {
							joined = append(joined, item.Title)
						}
						if !strings.Contains(strings.Join(joined, "|"), step.ExpectTitle) {
							t.Fatalf("step[%d] titles=%q want contains=%q", i, strings.Join(joined, "|"), step.ExpectTitle)
						}
						if len(resp.List) > 0 && resp.List[0].Finished != step.ExpectDone {
							t.Fatalf("step[%d] finished=%v want=%v", i, resp.List[0].Finished, step.ExpectDone)
						}
					}
				case "delete_item":
					l := logicpkg.NewDeleteHistoryLogic(context.Background(), svcCtx)
					if _, err := l.DeleteHistoryItem(&historyclient.DeleteHistoryItemRequest{UserID: step.UserID, MediaType: step.MediaType, MediaID: step.MediaID}); err != nil {
						t.Fatalf("step[%d] delete item error: %v", i, err)
					}
				case "clear_type":
					l := logicpkg.NewDeleteHistoryLogic(context.Background(), svcCtx)
					if _, err := l.ClearHistoryByType(&historyclient.ClearHistoryByTypeRequest{UserID: step.UserID, MediaType: step.MediaType}); err != nil {
						t.Fatalf("step[%d] clear type error: %v", i, err)
					}
				case "clear_all":
					l := logicpkg.NewDeleteHistoryLogic(context.Background(), svcCtx)
					if _, err := l.ClearHistoryAll(&historyclient.ClearHistoryAllRequest{UserID: step.UserID}); err != nil {
						t.Fatalf("step[%d] clear all error: %v", i, err)
					}
				default:
					t.Fatalf("step[%d] unsupported op=%s", i, step.Op)
				}
			}
		})
	}
}

func mustCreateHistoryTable(t *testing.T, db *sql.DB) {
	t.Helper()
	stmt := `
CREATE TABLE history_record (
  id bigint NOT NULL AUTO_INCREMENT,
  user_id bigint NOT NULL,
  media_type tinyint NOT NULL,
  media_id bigint NOT NULL,
  title varchar(255) NOT NULL DEFAULT '',
  cover varchar(1024) NOT NULL DEFAULT '',
  author_id bigint NOT NULL DEFAULT 0,
  progress_ms bigint NOT NULL DEFAULT 0,
  duration_ms bigint NOT NULL DEFAULT 0,
  finished tinyint NOT NULL DEFAULT 0,
  source tinyint NOT NULL DEFAULT 0,
  device varchar(255) NOT NULL DEFAULT '',
  meta text NOT NULL,
  first_seen_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  last_seen_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  deleted tinyint NOT NULL DEFAULT 0,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_media_active (user_id, media_type, media_id, deleted),
  KEY idx_user_last_seen (user_id, deleted, last_seen_at, id),
  KEY idx_user_type_last_seen (user_id, media_type, deleted, last_seen_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	if _, err := db.Exec(stmt); err != nil {
		t.Fatalf("create history table: %v", err)
	}
}
