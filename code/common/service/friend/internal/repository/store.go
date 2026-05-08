package repository

import "context"

type FriendRequest struct {
	ID         int64
	FromUserID int64
	ToUserID   int64
	Domain     string
	TenantID   string
	Message    string
	Status     int32
	CreatedAt  int64
	UpdatedAt  int64
}

type FriendStore interface {
	AddFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) error
	RemoveFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) error
	ListFriends(ctx context.Context, userID int64, domain, tenantID string, offset, limit int) ([]int64, int64, error)
	IsFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) (bool, error)

	CreateRequest(ctx context.Context, req *FriendRequest) (int64, error)
	GetRequest(ctx context.Context, requestID int64) (*FriendRequest, error)
	ListIncomingRequests(ctx context.Context, userID int64, status int32, offset, limit int) ([]*FriendRequest, int64, error)
	UpdateRequestStatus(ctx context.Context, requestID int64, status int32) error

	BlockUser(ctx context.Context, userID, blockedUserID int64, domain, tenantID string) error
	UnblockUser(ctx context.Context, userID, blockedUserID int64, domain, tenantID string) error
	IsBlocked(ctx context.Context, userID, targetUserID int64, domain, tenantID string) (bool, error)
	ListBlockedUsers(ctx context.Context, userID int64, domain, tenantID string, offset, limit int) ([]int64, int64, error)
}
