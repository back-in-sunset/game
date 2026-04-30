package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	MediaTypePost  = 1
	MediaTypeVideo = 2

	DefaultPageSize = 20
	MaxPageSize     = 100
)

type HistoryRecord struct {
	ID          int64        `db:"id"`
	UserID      int64        `db:"user_id"`
	MediaType   int64        `db:"media_type"`
	MediaID     int64        `db:"media_id"`
	Title       string       `db:"title"`
	Cover       string       `db:"cover"`
	AuthorID    int64        `db:"author_id"`
	ProgressMs  int64        `db:"progress_ms"`
	DurationMs  int64        `db:"duration_ms"`
	Finished    int64        `db:"finished"`
	Source      int64        `db:"source"`
	Device      string       `db:"device"`
	Meta        string       `db:"meta"`
	FirstSeenAt sql.NullTime `db:"first_seen_at"`
	LastSeenAt  sql.NullTime `db:"last_seen_at"`
	Deleted     int64        `db:"deleted"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

type HistoryModel interface {
	UpsertRecord(ctx context.Context, data *HistoryRecord) (*HistoryRecord, error)
	ListByUser(ctx context.Context, userID int64, mediaType int64, cursor int64, lastID int64, pageSize int64) ([]*HistoryRecord, error)
	SoftDeleteItem(ctx context.Context, userID int64, mediaType int64, mediaID int64) error
	SoftDeleteByType(ctx context.Context, userID int64, mediaType int64) error
	SoftDeleteAll(ctx context.Context, userID int64) error
	PurgeExpired(ctx context.Context, before time.Time) (int64, error)
}

type mysqlHistoryModel struct {
	conn  sqlx.SqlConn
	table string
}

const (
	historyRecordColumns = "id,user_id,media_type,media_id,title,cover,author_id,progress_ms,duration_ms,finished,source,device,meta,first_seen_at,last_seen_at,deleted,created_at,updated_at"
)

func NewHistoryModel(conn sqlx.SqlConn) HistoryModel {
	return &mysqlHistoryModel{conn: conn, table: "`history_record`"}
}

func (m *mysqlHistoryModel) UpsertRecord(ctx context.Context, data *HistoryRecord) (*HistoryRecord, error) {
	if data.ProgressMs < 0 || data.DurationMs < 0 {
		return nil, ErrInvalidProgress
	}
	if data.DurationMs > 0 && data.Finished == 0 && data.ProgressMs > data.DurationMs {
		return nil, ErrInvalidProgress
	}

	now := time.Now()
	query := fmt.Sprintf("insert into %s (user_id,media_type,media_id,title,cover,author_id,progress_ms,duration_ms,finished,source,device,meta,first_seen_at,last_seen_at,deleted) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,0) on duplicate key update id=last_insert_id(id),title=values(title),cover=values(cover),author_id=values(author_id),progress_ms=values(progress_ms),duration_ms=values(duration_ms),finished=values(finished),source=values(source),device=values(device),meta=values(meta),last_seen_at=values(last_seen_at),deleted=0", m.table)
	res, err := m.conn.ExecCtx(ctx, query, data.UserID, data.MediaType, data.MediaID, data.Title, data.Cover, data.AuthorID, data.ProgressMs, data.DurationMs, data.Finished, data.Source, data.Device, data.Meta, now, now)
	if err != nil {
		return nil, err
	}
	data.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return m.findOne(ctx, data.ID)
}

func (m *mysqlHistoryModel) ListByUser(ctx context.Context, userID int64, mediaType int64, cursor int64, lastID int64, pageSize int64) ([]*HistoryRecord, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	args := []any{userID}
	where := "where user_id=? and deleted=0"
	if mediaType > 0 {
		where += " and media_type=?"
		args = append(args, mediaType)
	}
	if cursor > 0 {
		cursorTime := time.Unix(cursor, 0)
		where += " and (last_seen_at < ? or (last_seen_at = ? and id < ?))"
		args = append(args, cursorTime, cursorTime, lastID)
	}
	args = append(args, pageSize)

	query := "select " + historyRecordColumns + " from history_record " + where + " order by last_seen_at desc,id desc limit ?"
	var records []*HistoryRecord
	if err := m.conn.QueryRowsCtx(ctx, &records, query, args...); err != nil {
		return nil, err
	}
	return records, nil
}

func (m *mysqlHistoryModel) SoftDeleteItem(ctx context.Context, userID int64, mediaType int64, mediaID int64) error {
	_, err := m.conn.ExecCtx(ctx, "update history_record set deleted=1 where user_id=? and media_type=? and media_id=? and deleted=0", userID, mediaType, mediaID)
	return err
}

func (m *mysqlHistoryModel) SoftDeleteByType(ctx context.Context, userID int64, mediaType int64) error {
	_, err := m.conn.ExecCtx(ctx, "update history_record set deleted=1 where user_id=? and media_type=? and deleted=0", userID, mediaType)
	return err
}

func (m *mysqlHistoryModel) SoftDeleteAll(ctx context.Context, userID int64) error {
	_, err := m.conn.ExecCtx(ctx, "update history_record set deleted=1 where user_id=? and deleted=0", userID)
	return err
}

func (m *mysqlHistoryModel) PurgeExpired(ctx context.Context, before time.Time) (int64, error) {
	res, err := m.conn.ExecCtx(ctx, "delete from history_record where last_seen_at < ?", before)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (m *mysqlHistoryModel) findOne(ctx context.Context, id int64) (*HistoryRecord, error) {
	var record HistoryRecord
	query := "select " + historyRecordColumns + " from history_record where id=? limit 1"
	err := m.conn.QueryRowCtx(ctx, &record, query, id)
	if err != nil {
		if err == sqlc.ErrNotFound || err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &record, nil
}
