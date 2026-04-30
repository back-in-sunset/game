package historycache

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"game/server/core/testkit"
	"history/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type stubHistoryModel struct {
	records      map[string]*model.HistoryRecord
	upsertCount  int
	deleteCount  int
	clearTypeCnt int
	clearAllCnt  int
}

func newStubHistoryModel() *stubHistoryModel {
	return &stubHistoryModel{records: make(map[string]*model.HistoryRecord)}
}

func (m *stubHistoryModel) UpsertRecord(_ context.Context, data *model.HistoryRecord) (*model.HistoryRecord, error) {
	m.upsertCount++
	key := historyIdentity(data.MediaType, data.MediaID)
	cloned := cloneHistoryRecord(data)
	if cloned.ID == 0 {
		cloned.ID = int64(m.upsertCount)
	}
	m.records[key] = cloned
	return cloneHistoryRecord(cloned), nil
}

func (m *stubHistoryModel) ListByUser(_ context.Context, _ int64, _ int64, _ int64, _ int64, _ int64) ([]*model.HistoryRecord, error) {
	return nil, nil
}

func (m *stubHistoryModel) SoftDeleteItem(_ context.Context, _ int64, mediaType int64, mediaID int64) error {
	m.deleteCount++
	delete(m.records, historyIdentity(mediaType, mediaID))
	return nil
}

func (m *stubHistoryModel) SoftDeleteByType(_ context.Context, _ int64, mediaType int64) error {
	m.clearTypeCnt++
	for key, record := range m.records {
		if record.MediaType == mediaType {
			delete(m.records, key)
		}
	}
	return nil
}

func (m *stubHistoryModel) SoftDeleteAll(_ context.Context, _ int64) error {
	m.clearAllCnt++
	m.records = make(map[string]*model.HistoryRecord)
	return nil
}

func (m *stubHistoryModel) PurgeExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestManagerRecordListAndFlush(t *testing.T) {
	redisAddr := testkit.StartMiniRedis(t)
	rds, err := redis.NewRedis(redis.RedisConf{Host: redisAddr, Type: "node"})
	if err != nil {
		t.Fatalf("new redis: %v", err)
	}

	stub := newStubHistoryModel()
	manager := NewManager(stub, rds, Config{
		WriteBackEnabled: true,
		ReadFallbackToDB: true,
		FlushBatchUsers:  10,
		FlushBatchItems:  10,
	})

	ctx := context.Background()
	record, err := manager.Record(ctx, &model.HistoryRecord{
		UserID:     1001,
		MediaType:  model.MediaTypeVideo,
		MediaID:    3001,
		Title:      "video first",
		ProgressMs: 12,
		DurationMs: 34,
		Finished:   0,
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if record == nil || !record.LastSeenAt.Valid {
		t.Fatalf("invalid record: %+v", record)
	}

	result, err := manager.List(ctx, 1001, 0, 0, 0, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Records) != 1 || result.Records[0].Title != "video first" {
		t.Fatalf("unexpected list result: %+v", result)
	}

	if err := manager.FlushOnce(ctx); err != nil {
		t.Fatalf("flush once: %v", err)
	}
	if stub.upsertCount != 1 {
		t.Fatalf("upsert count=%d want=1", stub.upsertCount)
	}
}

func TestManagerDeleteAndClear(t *testing.T) {
	redisAddr := testkit.StartMiniRedis(t)
	rds, err := redis.NewRedis(redis.RedisConf{Host: redisAddr, Type: "node"})
	if err != nil {
		t.Fatalf("new redis: %v", err)
	}

	stub := newStubHistoryModel()
	manager := NewManager(stub, rds, Config{
		WriteBackEnabled: true,
		ReadFallbackToDB: true,
		FlushBatchUsers:  10,
		FlushBatchItems:  10,
	})

	ctx := context.Background()
	for _, record := range []*model.HistoryRecord{
		{UserID: 1001, MediaType: model.MediaTypePost, MediaID: 1, Title: "post"},
		{UserID: 1001, MediaType: model.MediaTypeVideo, MediaID: 2, Title: "video"},
	} {
		if _, err := manager.Record(ctx, record); err != nil {
			t.Fatalf("seed record: %v", err)
		}
	}

	if err := manager.DeleteItem(ctx, 1001, model.MediaTypeVideo, 2); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	if err := manager.FlushOnce(ctx); err != nil {
		t.Fatalf("flush delete: %v", err)
	}
	if stub.deleteCount != 1 {
		t.Fatalf("delete count=%d want=1", stub.deleteCount)
	}

	if err := manager.ClearByType(ctx, 1001, model.MediaTypePost); err != nil {
		t.Fatalf("clear type: %v", err)
	}
	if err := manager.FlushOnce(ctx); err != nil {
		t.Fatalf("flush clear type: %v", err)
	}
	if stub.clearTypeCnt != 1 {
		t.Fatalf("clearType count=%d want=1", stub.clearTypeCnt)
	}

	if _, err := manager.Record(ctx, &model.HistoryRecord{UserID: 1001, MediaType: model.MediaTypePost, MediaID: 3, Title: "new"}); err != nil {
		t.Fatalf("record before clear all: %v", err)
	}
	if err := manager.ClearAll(ctx, 1001); err != nil {
		t.Fatalf("clear all: %v", err)
	}
	if err := manager.FlushOnce(ctx); err != nil {
		t.Fatalf("flush clear all: %v", err)
	}
	if stub.clearAllCnt != 1 {
		t.Fatalf("clearAll count=%d want=1", stub.clearAllCnt)
	}
}

func TestCloneHistoryRecordNil(t *testing.T) {
	if got := cloneHistoryRecord(nil); got != nil {
		t.Fatalf("clone nil = %+v, want nil", got)
	}
}

var _ model.HistoryModel = (*stubHistoryModel)(nil)
var _ = sql.ErrNoRows
