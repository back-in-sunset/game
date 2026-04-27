package logic

import (
	"context"
	"fmt"
	"strconv"

	"comment/rpc/internal/svc"
	"comment/rpc/types"
)

const (
	prefixCommentLikedUsers = "biz#comment:liked:users:objID:%d:commentID:%d"
)

const likeCommentLua = `
local likedUsersKey = KEYS[1]
local likeZsetKey = KEYS[2]
local likeZsetCompatKey = KEYS[3]
local memberID = ARGV[1]
local commentID = ARGV[2]

local added = redis.call("SADD", likedUsersKey, memberID)
if tonumber(added) == 1 then
	local score = redis.call("ZINCRBY", likeZsetKey, 1, commentID)
	redis.call("ZINCRBY", likeZsetCompatKey, 1, commentID)
	return {1, score}
end

local current = redis.call("ZSCORE", likeZsetKey, commentID)
if not current then
	current = "0"
end
return {0, current}
`

const unLikeCommentLua = `
local likedUsersKey = KEYS[1]
local likeZsetKey = KEYS[2]
local likeZsetCompatKey = KEYS[3]
local memberID = ARGV[1]
local commentID = ARGV[2]

local removed = redis.call("SREM", likedUsersKey, memberID)
if tonumber(removed) == 1 then
	local current = redis.call("ZSCORE", likeZsetKey, commentID)
	if not current then
		current = "0"
	end

	local score = tonumber(current)
	if score <= 0 then
		redis.call("ZADD", likeZsetKey, 0, commentID)
		redis.call("ZADD", likeZsetCompatKey, 0, commentID)
		score = 0
	else
		score = tonumber(redis.call("ZINCRBY", likeZsetKey, -1, commentID))
		redis.call("ZINCRBY", likeZsetCompatKey, -1, commentID)
		if score < 0 then
			redis.call("ZADD", likeZsetKey, 0, commentID)
			redis.call("ZADD", likeZsetCompatKey, 0, commentID)
			score = 0
		end
	end

	return {1, tostring(score)}
end

local current = redis.call("ZSCORE", likeZsetKey, commentID)
if not current then
	current = "0"
end
return {0, current}
`

type likeStoreResult struct {
	Changed   bool
	LikeCount int64
}

func likeKeyLikedUsers(objID int64, commentID int64) string {
	return fmt.Sprintf(prefixCommentLikedUsers, objID, commentID)
}

func likeKeyByLikeScore(objID int64, objType int64, rootID int64) string {
	return formatCommentIDsKey(objID, objType, rootID, types.SortLikeCount)
}

func likeKeyByLikeScoreCompat(objID int64, objType int64, rootID int64) string {
	return formatCommentObjSortTypeKey(objID, objType, rootID, types.SortLikeCount)
}

func execLikeScript(ctx context.Context, svcCtx *svc.ServiceContext, inObjID int64, inObjType int64, inRootID int64, inCommentID int64, inMemberID int64) (*likeStoreResult, error) {
	return execLikeMutationScript(ctx, svcCtx, likeCommentLua, inObjID, inObjType, inRootID, inCommentID, inMemberID)
}

func execUnlikeScript(ctx context.Context, svcCtx *svc.ServiceContext, inObjID int64, inObjType int64, inRootID int64, inCommentID int64, inMemberID int64) (*likeStoreResult, error) {
	return execLikeMutationScript(ctx, svcCtx, unLikeCommentLua, inObjID, inObjType, inRootID, inCommentID, inMemberID)
}

func execLikeMutationScript(ctx context.Context, svcCtx *svc.ServiceContext, script string, inObjID int64, inObjType int64, inRootID int64, inCommentID int64, inMemberID int64) (*likeStoreResult, error) {
	res, err := svcCtx.BizRedis.EvalCtx(ctx, script, []string{
		likeKeyLikedUsers(inObjID, inCommentID),
		likeKeyByLikeScore(inObjID, inObjType, inRootID),
		likeKeyByLikeScoreCompat(inObjID, inObjType, inRootID),
	},
		strconv.FormatInt(inMemberID, 10),
		strconv.FormatInt(inCommentID, 10),
	)
	if err != nil {
		return nil, err
	}

	items, ok := res.([]any)
	if !ok || len(items) < 2 {
		return nil, fmt.Errorf("unexpected redis eval result: %T %v", res, res)
	}

	changed, err := castToInt64(items[0])
	if err != nil {
		return nil, err
	}
	likeCount, err := castToInt64(items[1])
	if err != nil {
		return nil, err
	}

	if likeCount < 0 {
		likeCount = 0
	}
	return &likeStoreResult{
		Changed:   changed == 1,
		LikeCount: likeCount,
	}, nil
}

func castToInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	case float64:
		return int64(x), nil
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		if err == nil {
			return i, nil
		}
		f, ferr := strconv.ParseFloat(x, 64)
		if ferr != nil {
			return 0, err
		}
		return int64(f), nil
	case []byte:
		return castToInt64(string(x))
	default:
		return 0, fmt.Errorf("unsupported number type: %T", v)
	}
}
