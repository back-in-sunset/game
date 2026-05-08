package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RelationshipStore struct {
	conn sqlx.SqlConn
}

func NewRelationshipStore(conn sqlx.SqlConn) *RelationshipStore {
	return &RelationshipStore{conn: conn}
}

func (s *RelationshipStore) AddFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) error {
	id := genID()
	now := time.Now().UnixMilli()
	_, err := s.conn.ExecCtx(ctx, `
		INSERT INTO friend_relationship (id, user_id, friend_id, domain, tenant_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, userID, friendID, domain, tenantID, now)
	return err
}

func (s *RelationshipStore) RemoveFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) error {
	return s.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		_, err := session.ExecCtx(ctx, `
			DELETE FROM friend_relationship
			WHERE user_id = ? AND friend_id = ? AND domain = ? AND tenant_id = ?
		`, userID, friendID, domain, tenantID)
		if err != nil {
			return err
		}
		_, err = session.ExecCtx(ctx, `
			DELETE FROM friend_relationship
			WHERE user_id = ? AND friend_id = ? AND domain = ? AND tenant_id = ?
		`, friendID, userID, domain, tenantID)
		return err
	})
}

func (s *RelationshipStore) ListFriends(ctx context.Context, userID int64, domain, tenantID string, offset, limit int) ([]int64, int64, error) {
	var total int64
	err := s.conn.QueryRowCtx(ctx, &total, `
		SELECT COUNT(*) FROM friend_relationship
		WHERE user_id = ? AND domain = ? AND tenant_id = ?
	`, userID, domain, tenantID)
	if err != nil {
		return nil, 0, err
	}

	var friends []int64
	err = s.conn.QueryRowsCtx(ctx, &friends, `
		SELECT friend_id FROM friend_relationship
		WHERE user_id = ? AND domain = ? AND tenant_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, userID, domain, tenantID, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []int64{}, 0, nil
		}
		return nil, 0, err
	}

	if friends == nil {
		friends = []int64{}
	}
	return friends, total, nil
}

func (s *RelationshipStore) IsFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) (bool, error) {
	var count int64
	err := s.conn.QueryRowCtx(ctx, &count, `
		SELECT COUNT(*) FROM friend_relationship
		WHERE user_id = ? AND friend_id = ? AND domain = ? AND tenant_id = ?
	`, userID, friendID, domain, tenantID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

var idSeq int64

func genID() int64 {
	idSeq++
	return time.Now().UnixNano() + idSeq
}
