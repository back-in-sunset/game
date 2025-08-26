package logic

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"comment/internal/errx"
	"comment/model"
	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/types"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

const (
	prefixCommentIDs = "biz#commentids#objID:%d:objType:%d:sortType:%d"
	commentIDsExpire = 3600 * 24 * 2
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

// GetCommentList 获取评论列表
func (l *GetCommentListLogic) GetCommentList(in *comment.CommentListRequest) (*comment.CommentListResponse, error) {
	var (
		err            error
		isCache, isEnd bool
		lastID         int64
		curPage        []*comment.CommentResponse
		comments       []*model.Comment
		sortFiled      string
	)

	cursor := in.Cursor

	switch in.SortType {
	case types.SortLikeCount:
		// 按点赞数排序
		sortFiled = "like_count"
		// 分页游标
		if cursor <= 0 {
			cursor = types.DefaultSortLikeCursor
		}
	case types.SortCreatedTime:
		// 按创建时间排序
		sortFiled = "created_at"
		// 分页游标
		if cursor <= 0 {
			cursor = time.Now().Unix()
		}
	default:
		return nil, errx.ErrorSortTypeInvalid
	}

	// 分页大小
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = types.DefaultPageSize
	}

	commentIDs, _ := l.cacheCommentIDs(in.ObjID, in.ObjType, cursor, pageSize, in.SortType)
	if len(commentIDs) > 0 {
		isCache = true
		if commentIDs[len(commentIDs)-1] == -1 {
			isEnd = true
		}

		comments, err = l.svcCtx.CommentModel.CacheCommentsByIDs(l.ctx, in.ObjID, commentIDs)
		if err != nil {
			return nil, err
		}

		// 排序
		var cmpfunc func(a, b *model.Comment) int
		switch in.SortType {
		case types.SortLikeCount:
			// 按点赞数排序
			cmpfunc = func(a, b *model.Comment) int {
				return cmp.Compare(b.LikeCount, a.LikeCount)
			}
		case types.SortCreatedTime:
			// 按创建时间排序
			cmpfunc = func(a, b *model.Comment) int {
				return cmp.Compare(b.CreatedAt.Unix(), a.CreatedAt.Unix())
			}
		}
		slices.SortFunc(comments, cmpfunc)

		for _, c := range comments {
			var commentResponse comment.CommentResponse
			copier.Copy(&commentResponse, c)
			curPage = append(curPage, &commentResponse)
		}

	} else {
		// 从数据库查询
		v, err, _ := l.svcCtx.SignleFlightGroup.Do("commentsByObjID", func() (any, error) {
			return l.svcCtx.CommentModel.CommentListByObjID(l.ctx, in.ObjID, in.ObjType, sortFiled, types.DefaultLimit)
		})
		if err != nil {
			return nil, err
		}
		comments = v.([]*model.Comment)
		var firstPageComments []*model.Comment
		if len(comments) > int(in.PageSize) {
			firstPageComments = comments[:in.PageSize]
		} else {
			firstPageComments = comments
			isEnd = true
		}

		for _, c := range firstPageComments {
			var comment comment.CommentResponse
			copier.Copy(&comment, c)
			curPage = append(curPage, &comment)
		}
	}

	if len(curPage) > 0 {
		pageLastRecord := curPage[len(curPage)-1]
		lastID = pageLastRecord.ID
		if in.SortType == types.SortCreatedTime {
			cursor = pageLastRecord.CreatedAt
		} else {
			cursor = pageLastRecord.LikeCount
		}

		if cursor < 0 {
			cursor = 0
		}

		for k, c := range curPage {
			if in.SortType == types.SortCreatedTime {
				if c.CreatedAt == in.Cursor && c.CommentID == in.CommentID {
					curPage = curPage[k:]
					break
				}
			} else {
				if c.LikeCount == in.Cursor && c.CommentID == in.CommentID {
					curPage = curPage[k:]
					break
				}
			}
		}

	}

	ret := &comment.CommentListResponse{
		IsEnd:    isEnd,
		Cursor:   cursor,
		Comments: curPage,
		LastID:   lastID,
	}

	if !isCache {
		threading.GoSafe(func() {
			if len(comments) < types.DefaultLimit && len(comments) > 0 {
				comments = append(comments, &model.Comment{ID: -1})
			}
			err = l.addComments(context.Background(), in.ObjID, in.ObjType, in.SortType, comments)
			if err != nil {
				l.Logger.Errorf("addComments error: %v", err)
			}

		})
	}

	return ret, nil
}

func (l *GetCommentListLogic) cacheCommentIDs(objID int64, objType int64, cursor int64, pageSize int64, sortType int64) ([]int64, error) {
	key := fmt.Sprintf(prefixCommentIDs, objID, objType, sortType)

	ok, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, key)
	if err != nil {
		l.Logger.Errorf("cacheCommentIDs error: %v", err)
	}
	if ok {
		err = l.svcCtx.BizRedis.ExpireCtx(l.ctx, key, commentIDsExpire)
		if err != nil {
			l.Logger.Errorf("cacheCommentIDs error: %v", err)
		}
	}

	pairs, err := l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(l.ctx, key, 0, cursor, 0, int(pageSize))
	if err != nil {
		l.Logger.Errorf("cacheCommentIDs error: %v", err)
		return nil, err
	}

	var ids []int64
	for _, pair := range pairs {
		id, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			l.Logger.Errorf("strconv commentid from redis err:%v", err)
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (l *GetCommentListLogic) addComments(ctx context.Context, objID int64, objType int64, sortType int64, comments []*model.Comment) error {
	if len(comments) == 0 {
		return nil
	}
	key := fmt.Sprintf(prefixCommentIDs, objID, objType, sortType)
	for _, c := range comments {
		var score int64
		if sortType == types.SortCreatedTime {
			score = c.CreatedAt.Unix()
		} else {
			score = c.LikeCount
		}

		if score <= 0 {
			score = 0
		}

		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.FormatInt(c.ID, 10))
		if err != nil {
			l.Logger.Errorf("addComments error: %v", err)
			return err
		}
	}

	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, commentIDsExpire)
}
