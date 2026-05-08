package mysql

import (
	"context"
	"database/sql"
	"time"

	"friend/internal/errx"
	"friend/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/status"
)

type RequestStore struct {
	conn sqlx.SqlConn
}

func NewRequestStore(conn sqlx.SqlConn) *RequestStore {
	return &RequestStore{conn: conn}
}

func (s *RequestStore) CreateRequest(ctx context.Context, req *repository.FriendRequest) (int64, error) {
	id := genID()
	now := time.Now().UnixMilli()
	_, err := s.conn.ExecCtx(ctx, `
		INSERT INTO friend_request (id, from_user_id, to_user_id, domain, tenant_id, message, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)
	`, id, req.FromUserID, req.ToUserID, req.Domain, req.TenantID, req.Message, now, now)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *RequestStore) GetRequest(ctx context.Context, requestID int64) (*repository.FriendRequest, error) {
	var req repository.FriendRequest
	err := s.conn.QueryRowCtx(ctx, &req, `
		SELECT id, from_user_id, to_user_id, domain, tenant_id, message, status, created_at, updated_at
		FROM friend_request WHERE id = ?
	`, requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(404, errx.F1004)
		}
		return nil, err
	}
	return &req, nil
}

func (s *RequestStore) ListIncomingRequests(ctx context.Context, userID int64, statusFilter int32, offset, limit int) ([]*repository.FriendRequest, int64, error) {
	var total int64
	args := []interface{}{userID, statusFilter}
	err := s.conn.QueryRowCtx(ctx, &total, `
		SELECT COUNT(*) FROM friend_request
		WHERE to_user_id = ? AND status = ?
	`, args...)
	if err != nil {
		return nil, 0, err
	}

	var reqs []*repository.FriendRequest
	err = s.conn.QueryRowsCtx(ctx, &reqs, `
		SELECT id, from_user_id, to_user_id, domain, tenant_id, message, status, created_at, updated_at
		FROM friend_request
		WHERE to_user_id = ? AND status = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, userID, statusFilter, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*repository.FriendRequest{}, 0, nil
		}
		return nil, 0, err
	}

	if reqs == nil {
		reqs = []*repository.FriendRequest{}
	}
	return reqs, total, nil
}

func (s *RequestStore) UpdateRequestStatus(ctx context.Context, requestID int64, newStatus int32) error {
	now := time.Now().UnixMilli()
	_, err := s.conn.ExecCtx(ctx, `
		UPDATE friend_request SET status = ?, updated_at = ? WHERE id = ?
	`, newStatus, now, requestID)
	return err
}
