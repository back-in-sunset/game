package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type BlockStore struct {
	conn sqlx.SqlConn
}

func NewBlockStore(conn sqlx.SqlConn) *BlockStore {
	return &BlockStore{conn: conn}
}

func (s *BlockStore) BlockUser(ctx context.Context, userID, blockedUserID int64, domain, tenantID string) error {
	id := genID()
	now := time.Now().UnixMilli()
	_, err := s.conn.ExecCtx(ctx, `
		INSERT INTO friend_block (id, user_id, blocked_user_id, domain, tenant_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, userID, blockedUserID, domain, tenantID, now)
	return err
}

func (s *BlockStore) UnblockUser(ctx context.Context, userID, blockedUserID int64, domain, tenantID string) error {
	_, err := s.conn.ExecCtx(ctx, `
		DELETE FROM friend_block
		WHERE user_id = ? AND blocked_user_id = ? AND domain = ? AND tenant_id = ?
	`, userID, blockedUserID, domain, tenantID)
	return err
}

func (s *BlockStore) IsBlocked(ctx context.Context, userID, targetUserID int64, domain, tenantID string) (bool, error) {
	var count int64
	err := s.conn.QueryRowCtx(ctx, &count, `
		SELECT COUNT(*) FROM friend_block
		WHERE user_id = ? AND blocked_user_id = ? AND domain = ? AND tenant_id = ?
	`, userID, targetUserID, domain, tenantID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *BlockStore) ListBlockedUsers(ctx context.Context, userID int64, domain, tenantID string, offset, limit int) ([]int64, int64, error) {
	var total int64
	err := s.conn.QueryRowCtx(ctx, &total, `
		SELECT COUNT(*) FROM friend_block
		WHERE user_id = ? AND domain = ? AND tenant_id = ?
	`, userID, domain, tenantID)
	if err != nil {
		return nil, 0, err
	}

	var blocked []int64
	err = s.conn.QueryRowsCtx(ctx, &blocked, `
		SELECT blocked_user_id FROM friend_block
		WHERE user_id = ? AND domain = ? AND tenant_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, userID, domain, tenantID, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []int64{}, 0, nil
		}
		return nil, 0, err
	}

	if blocked == nil {
		blocked = []int64{}
	}
	return blocked, total, nil
}
