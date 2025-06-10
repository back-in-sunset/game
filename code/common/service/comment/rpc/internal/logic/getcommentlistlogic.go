package logic

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"comment/model"
	"comment/rpc/internal/svc"
	"comment/rpc/internal/types"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
	"github.com/zeromicro/go-zero/core/threading"
)

type GetCommentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取评论列表
func (l *GetCommentListLogic) GetCommentList(in *comment.CommentListRequest) (*comment.CommentListResponse, error) {
	if in.SortType != types.SortPublishTime && in.SortType != types.SortLikeCount {
		return nil, fmt.Errorf("sortType is not valid")
	}
	if in.ObjId <= 0 {
		return nil, fmt.Errorf("bojId is not valid")
	}
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}
	if in.Cursor == 0 {
		if in.SortType == types.SortPublishTime {
			in.Cursor = uint64(time.Now().Unix())
		} else {
			in.Cursor = types.DefaultSortLikeCursor
		}
	}

	var (
		sortField       string
		sortLikeCount   uint64
		sortPublishTime string
	)
	if in.SortType == types.SortLikeCount {
		sortField = "like_num"
		sortLikeCount = in.Cursor
	} else {
		sortField = "publish_time"
		sortPublishTime = time.Unix(int64(in.Cursor), 0).Format("2006-01-02 15:04:05")
	}

	var (
		err            error
		isCache, isEnd bool
		lastId, cursor int64
		cachePage      []*comment.CommentResponse
		// model
		comments []*model.Comment
	)

	// 查询索引 再根据索引查询内容
	// 1. id cache aside
	// 2  id is last
	// 2  build cache

	commentIds, _ := l.cacheCommentIds(l.ctx, in.ObjId, in.ObjType, int64(in.Cursor), int64(in.PageSize), int64(in.SortType))
	if len(commentIds) > 0 {
		isCache = true
		if commentIds[len(commentIds)-1] == -1 {
			isEnd = true
		}
		comments, err = l.commentsByIds(l.ctx, in.ObjId, commentIds)
		if err != nil {
			return nil, err
		}

		// 通过sortFiled对comments进行排序
		var cmpFunc func(a, b *model.Comment) int
		if sortField == "like_num" {
			cmpFunc = func(a, b *model.Comment) int {
				return cmp.Compare(b.LikeCount, a.LikeCount)
			}
		} else {
			cmpFunc = func(a, b *model.Comment) int {
				return cmp.Compare(b.CreatedAt.Unix(), b.CreatedAt.Unix())
			}
		}
		slices.SortFunc(comments, cmpFunc)
		for _, commentModel := range comments {
			cachePage = append(cachePage, &comment.CommentResponse{
				ObjId:       commentModel.ObjId,
				ObjType:     commentModel.ObjType,
				MemberId:    commentModel.MemberId,
				CommentId:   uint64(commentModel.Id),
				AtMemberIds: commentModel.AtMemberIds,
				Ip:          commentModel.Ip,
				Platform:    commentModel.Platform,
				Device:      commentModel.Device,
				Message:     commentModel.Message,
				Meta:        commentModel.Meta,
				ReplyId:     commentModel.ReplyId,
				State:       commentModel.State,
				RootId:      commentModel.RootId,
				CreatedAt:   uint64(commentModel.CreatedAt.Unix()),
				Floor:       commentModel.Floor,
				LikeCount:   uint64(commentModel.LikeCount),
				HateCount:   uint64(commentModel.HateCount),
				Count:       uint64(commentModel.Count),
			})
		}

	} else {
		// 没命中
		v, err, _ := l.svcCtx.SingleFlightGroup.Do(fmt.Sprintf("CommentByObjIds:%d:%d", in.ObjId, in.SortType), func() (interface{}, error) {
			return l.svcCtx.CommentModel.CommentByObjId(l.ctx, in.ObjId, in.ObjType, sortLikeCount, sortPublishTime, sortField, types.DefaultLimit)
		})
		if err != nil {
			logx.Errorf("ArticlesByUserId userId: %d sortField: %s error: %v", in.ObjId, sortField, err)
			return nil, err
		}
		if v == nil {
			return &comment.CommentListResponse{}, nil
		}
		comments = v.([]*model.Comment)
		var firstPageComments []*model.Comment
		if len(comments) > int(in.PageSize) {
			firstPageComments = comments[:int(in.PageSize)]
		} else {
			firstPageComments = comments
			isEnd = true
		}
		for _, commentModel := range firstPageComments {
			cachePage = append(cachePage, &comment.CommentResponse{
				ObjId:       commentModel.ObjId,
				ObjType:     commentModel.ObjType,
				MemberId:    commentModel.MemberId,
				CommentId:   uint64(commentModel.Id),
				AtMemberIds: commentModel.AtMemberIds,
				Ip:          commentModel.Ip,
				Platform:    commentModel.Platform,
				Device:      commentModel.Device,
				Message:     commentModel.Message,
				Meta:        commentModel.Meta,
				ReplyId:     commentModel.ReplyId,
				State:       commentModel.State,
				RootId:      commentModel.RootId,
				CreatedAt:   uint64(commentModel.CreatedAt.Unix()),
				Floor:       commentModel.Floor,
				LikeCount:   uint64(commentModel.LikeCount),
				HateCount:   uint64(commentModel.HateCount),
				Count:       uint64(commentModel.Count),
			})
		}
	}

	if len(cachePage) > 0 {
		pageLast := cachePage[len(cachePage)-1]
		lastId = int64(pageLast.CommentId)
		if in.SortType == types.SortPublishTime {
			cursor = int64(pageLast.CreatedAt)
		} else {
			cursor = int64(pageLast.LikeCount)
		}
		if cursor < 0 {
			cursor = 0
		}

		// 从开始offset开始截断成当前页
		for k, comment := range cachePage {
			if in.SortType == types.SortPublishTime {
				if comment.CreatedAt == in.Cursor && comment.CommentId == in.CommentId {
					cachePage = cachePage[k:]
					break
				}
			} else {
				if comment.LikeCount == in.Cursor && comment.CommentId == in.CommentId {
					cachePage = cachePage[k:]
					break
				}
			}
		}
	}

	ret := &comment.CommentListResponse{
		Cursor:   cursor,
		IsEnd:    isEnd,
		LastId:   lastId,
		Comments: cachePage,
	}

	if !isCache {
		threading.GoSafe(func() {
			if len(comments) < types.DefaultLimit && len(comments) > 0 {
				comments = append(comments, &model.Comment{Id: -1})
			}
			err = l.addCacheComments(context.Background(), comments, in.ObjId, in.ObjType, in.SortType)
			if err != nil {
				logx.Errorf("addCacheArticles error: %v", err)
			}
		})
	}

	return ret, nil
}

func (l *GetCommentListLogic) cacheCommentIds(ctx context.Context, objId, objType uint64, cursor, ps, sortType int64) ([]int64, error) {
	key := commentIdsKey(objId, objType, sortType)
	b, err := l.svcCtx.BizRedis.ExistsCtx(ctx, key)
	if err != nil {
		logx.Errorf("ExistsCtx key: %s error: %v", key, err)
	}
	if b {
		err = l.svcCtx.BizRedis.ExpireCtx(ctx, key, commentsExpire)
		if err != nil {
			logx.Errorf("ExpireCtx key: %s error: %v", key, err)
		}
	}

	pairs, err := l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, 0, cursor, 0, int(ps))
	if err != nil {
		logx.Errorf("ZrevrangebyscoreWithScoresAndLimit key: %s error: %v", key, err)
		return nil, err
	}
	var ids []int64
	for _, pair := range pairs {
		id, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			logx.Errorf("strconv.ParseInt key: %s error: %v", pair.Key, err)
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (l *GetCommentListLogic) commentsByIds(ctx context.Context, objId uint64, commentIds []int64) ([]*model.Comment, error) {
	type commentObjIds struct {
		Id    uint64
		ObjId uint64
	}

	return mr.MapReduce(func(source chan<- commentObjIds) {
		for _, cid := range commentIds {
			if cid == -1 {
				continue
			}
			source <- commentObjIds{
				Id:    uint64(cid),
				ObjId: objId,
			}
		}
	}, func(objIdid commentObjIds, writer mr.Writer[*model.Comment], cancel func(error)) {
		p, err := l.svcCtx.CommentModel.FindOneByObjId(ctx, objIdid.ObjId, objIdid.Id)
		if err != nil {
			cancel(err)
			return
		}
		writer.Write(p)
	}, func(pipe <-chan *model.Comment, writer mr.Writer[[]*model.Comment], cancel func(error)) {
		var comments []*model.Comment
		for comment := range pipe {
			comments = append(comments, comment)
		}
		writer.Write(comments)
	})
}
