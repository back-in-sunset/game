package mysql

import (
	"context"

	"friend/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FriendStore struct {
	rel   *RelationshipStore
	req   *RequestStore
	block *BlockStore
}

func NewFriendStore(conn sqlx.SqlConn) repository.FriendStore {
	return &FriendStore{
		rel:   NewRelationshipStore(conn),
		req:   NewRequestStore(conn),
		block: NewBlockStore(conn),
	}
}

// Relationship

func (s *FriendStore) AddFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) error {
	return s.rel.AddFriend(ctx, userID, friendID, domain, tenantID)
}

func (s *FriendStore) RemoveFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) error {
	return s.rel.RemoveFriend(ctx, userID, friendID, domain, tenantID)
}

func (s *FriendStore) ListFriends(ctx context.Context, userID int64, domain, tenantID string, offset, limit int) ([]int64, int64, error) {
	return s.rel.ListFriends(ctx, userID, domain, tenantID, offset, limit)
}

func (s *FriendStore) IsFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) (bool, error) {
	return s.rel.IsFriend(ctx, userID, friendID, domain, tenantID)
}

// Request

func (s *FriendStore) CreateRequest(ctx context.Context, req *repository.FriendRequest) (int64, error) {
	return s.req.CreateRequest(ctx, req)
}

func (s *FriendStore) GetRequest(ctx context.Context, requestID int64) (*repository.FriendRequest, error) {
	return s.req.GetRequest(ctx, requestID)
}

func (s *FriendStore) ListIncomingRequests(ctx context.Context, userID int64, status int32, offset, limit int) ([]*repository.FriendRequest, int64, error) {
	return s.req.ListIncomingRequests(ctx, userID, status, offset, limit)
}

func (s *FriendStore) UpdateRequestStatus(ctx context.Context, requestID int64, status int32) error {
	return s.req.UpdateRequestStatus(ctx, requestID, status)
}

// Block

func (s *FriendStore) BlockUser(ctx context.Context, userID, blockedUserID int64, domain, tenantID string) error {
	return s.block.BlockUser(ctx, userID, blockedUserID, domain, tenantID)
}

func (s *FriendStore) UnblockUser(ctx context.Context, userID, blockedUserID int64, domain, tenantID string) error {
	return s.block.UnblockUser(ctx, userID, blockedUserID, domain, tenantID)
}

func (s *FriendStore) IsBlocked(ctx context.Context, userID, targetUserID int64, domain, tenantID string) (bool, error) {
	return s.block.IsBlocked(ctx, userID, targetUserID, domain, tenantID)
}

func (s *FriendStore) ListBlockedUsers(ctx context.Context, userID int64, domain, tenantID string, offset, limit int) ([]int64, int64, error) {
	return s.block.ListBlockedUsers(ctx, userID, domain, tenantID, offset, limit)
}
