package friendcache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Cache struct {
	rds             *redis.Redis
	listTTLSeconds  int
	checkTTLSeconds int
}

func NewCache(rds *redis.Redis, listTTL, checkTTL int) *Cache {
	return &Cache{rds: rds, listTTLSeconds: listTTL, checkTTLSeconds: checkTTL}
}

func (c *Cache) listKey(userID int64, domain, tenantID string) string {
	return fmt.Sprintf("friend:list:%d:%s:%s", userID, domain, tenantID)
}

func (c *Cache) checkKey(userID, friendID int64, domain, tenantID string) string {
	return fmt.Sprintf("friend:check:%d:%d:%s:%s", userID, friendID, domain, tenantID)
}

func (c *Cache) blockedKey(userID, targetID int64, domain, tenantID string) string {
	return fmt.Sprintf("friend:blocked:%d:%d:%s:%s", userID, targetID, domain, tenantID)
}

func (c *Cache) ListFriends(ctx context.Context, userID int64, domain, tenantID string) ([]int64, bool) {
	key := c.listKey(userID, domain, tenantID)
	vals, err := c.rds.ZrevrangeWithScoresCtx(ctx, key, 0, -1)
	if err != nil || len(vals) == 0 {
		return nil, false
	}
	ids := make([]int64, 0, len(vals))
	for _, v := range vals {
		id, parseErr := strconv.ParseInt(v.Key, 10, 64)
		if parseErr != nil {
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
}

func (c *Cache) SetFriendList(ctx context.Context, userID int64, domain, tenantID string, friends []int64) error {
	key := c.listKey(userID, domain, tenantID)
	pairs := make([]redis.Pair, 0, len(friends))
	for i, f := range friends {
		pairs = append(pairs, redis.Pair{Key: strconv.FormatInt(f, 10), Score: int64(len(friends) - i)})
	}
	if _, err := c.rds.ZaddsCtx(ctx, key, pairs...); err != nil {
		return err
	}
	return c.rds.ExpireCtx(ctx, key, c.listTTLSeconds)
}

func (c *Cache) AddToList(ctx context.Context, userID, friendID int64, domain, tenantID string, score int64) error {
	key := c.listKey(userID, domain, tenantID)
	_, err := c.rds.ZaddCtx(ctx, key, score, strconv.FormatInt(friendID, 10))
	return err
}

func (c *Cache) RemoveFromList(ctx context.Context, userID, friendID int64, domain, tenantID string) error {
	key := c.listKey(userID, domain, tenantID)
	_, err := c.rds.ZremCtx(ctx, key, strconv.FormatInt(friendID, 10))
	return err
}

func (c *Cache) IsFriend(ctx context.Context, userID, friendID int64, domain, tenantID string) (bool, bool) {
	key := c.checkKey(userID, friendID, domain, tenantID)
	val, err := c.rds.GetCtx(ctx, key)
	if err != nil {
		return false, false
	}
	return val == "1", true
}

func (c *Cache) SetFriendCheck(ctx context.Context, userID, friendID int64, domain, tenantID string, isFriend bool) error {
	key := c.checkKey(userID, friendID, domain, tenantID)
	val := "0"
	if isFriend {
		val = "1"
	}
	if err := c.rds.SetCtx(ctx, key, val); err != nil {
		return err
	}
	return c.rds.ExpireCtx(ctx, key, c.checkTTLSeconds)
}

func (c *Cache) DelFriendCheck(ctx context.Context, userID, friendID int64, domain, tenantID string) error {
	key := c.checkKey(userID, friendID, domain, tenantID)
	_, err := c.rds.DelCtx(ctx, key)
	return err
}

func (c *Cache) IsBlocked(ctx context.Context, userID, targetID int64, domain, tenantID string) (bool, bool) {
	key := c.blockedKey(userID, targetID, domain, tenantID)
	val, err := c.rds.GetCtx(ctx, key)
	if err != nil {
		return false, false
	}
	return val == "1", true
}

func (c *Cache) SetBlockedCheck(ctx context.Context, userID, targetID int64, domain, tenantID string, blocked bool) error {
	key := c.blockedKey(userID, targetID, domain, tenantID)
	val := "0"
	if blocked {
		val = "1"
	}
	if err := c.rds.SetCtx(ctx, key, val); err != nil {
		return err
	}
	return c.rds.ExpireCtx(ctx, key, c.checkTTLSeconds)
}

func (c *Cache) DelBlockedCheck(ctx context.Context, userID, targetID int64, domain, tenantID string) error {
	key := c.blockedKey(userID, targetID, domain, tenantID)
	_, err := c.rds.DelCtx(ctx, key)
	return err
}
