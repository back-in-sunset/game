package logic

import (
	"comment/model"
	"comment/rpc/internal/types"
	"context"
	"fmt"
	"strconv"
)

const (
	prefixCommentIds = "biz#commentids#objid#objtype#%d#%d#%d"
	commentsExpire   = 3600 * 24 * 2
)

func commentIdsKey(objId, objType uint64, sortType int64) string {
	return fmt.Sprintf(prefixCommentIds, objId, objType, sortType)
}

func (l *GetCommentListLogic) addCacheComments(ctx context.Context, comments []*model.Comment, objId, objType uint64, sortType uint64) error {
	if len(comments) == 0 {
		return nil
	}
	key := commentIdsKey(objId, objType, int64(sortType))
	for _, comment := range comments {
		var score int64
		if sortType == types.SortLikeCount {
			score = comment.LikeCount
		} else if sortType == types.SortPublishTime && comment.Id != -1 {
			score = comment.CreatedAt.Local().Unix()
		}
		if score < 0 {
			score = 0
		}
		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.Itoa(int(comment.Id)))
		if err != nil {
			return err
		}
	}

	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, commentsExpire)
}
