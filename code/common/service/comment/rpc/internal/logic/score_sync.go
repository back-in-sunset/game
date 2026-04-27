package logic

import (
	"context"
	"strconv"

	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"comment/rpc/types"
)

const (
	attrsScoreShiftBits = 40
	attrsPinnedBit      = 1 << 0
)

func buildCompositeScore(attrs int64, base int64) int64 {
	if attrs < 0 {
		attrs = 0
	}
	if base < 0 {
		base = 0
	}

	const baseMask = int64((1 << attrsScoreShiftBits) - 1)
	return ((attrs & 0xFFFFF) << attrsScoreShiftBits) + (base & baseMask)
}

func updateCommentScoresForSort(ctx context.Context, svcCtx *svc.ServiceContext, objID int64, objType int64, rootID int64, commentID int64, sortType int64, score int64) error {
	if svcCtx == nil || svcCtx.BizRedis == nil {
		return nil
	}

	member := strconv.FormatInt(commentID, 10)
	keyPrimary := likeKeyBySort(objID, objType, rootID, sortType)
	keyCompat := likeKeyBySortCompat(objID, objType, rootID, sortType)
	if _, err := svcCtx.BizRedis.ZaddCtx(ctx, keyPrimary, score, member); err != nil {
		return err
	}
	if _, err := svcCtx.BizRedis.ZaddCtx(ctx, keyCompat, score, member); err != nil {
		return err
	}
	return nil
}

func removeCommentScores(ctx context.Context, svcCtx *svc.ServiceContext, objID int64, objType int64, rootID int64, commentID int64) error {
	if svcCtx == nil || svcCtx.BizRedis == nil {
		return nil
	}

	member := strconv.FormatInt(commentID, 10)
	keys := []string{
		likeKeyBySort(objID, objType, rootID, types.SortLikeCount),
		likeKeyBySortCompat(objID, objType, rootID, types.SortLikeCount),
		likeKeyBySort(objID, objType, rootID, types.SortCreatedTime),
		likeKeyBySortCompat(objID, objType, rootID, types.SortCreatedTime),
	}
	for _, key := range keys {
		if _, err := svcCtx.BizRedis.ZremCtx(ctx, key, member); err != nil {
			return err
		}
	}
	return nil
}

func syncCommentScores(ctx context.Context, svcCtx *svc.ServiceContext, c *model.Comment) error {
	if c == nil {
		return nil
	}

	if c.State != 0 {
		return removeCommentScores(ctx, svcCtx, c.ObjID, c.ObjType, c.RootID, c.ID)
	}

	likeScore := buildCompositeScore(c.Attrs, c.LikeCount)
	timeScore := buildCompositeScore(c.Attrs, c.CreatedAt.Unix())
	if err := updateCommentScoresForSort(ctx, svcCtx, c.ObjID, c.ObjType, c.RootID, c.ID, types.SortLikeCount, likeScore); err != nil {
		return err
	}
	if err := updateCommentScoresForSort(ctx, svcCtx, c.ObjID, c.ObjType, c.RootID, c.ID, types.SortCreatedTime, timeScore); err != nil {
		return err
	}
	return nil
}

func likeKeyBySort(objID int64, objType int64, rootID int64, sortType int64) string {
	return formatCommentIDsKey(objID, objType, rootID, sortType)
}

func likeKeyBySortCompat(objID int64, objType int64, rootID int64, sortType int64) string {
	return formatCommentObjSortTypeKey(objID, objType, rootID, sortType)
}
