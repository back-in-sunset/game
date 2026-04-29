package logic

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"comment/internal/errx"
	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"comment/rpc/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

// 只处理id缓存 内容缓存交给model
const (
	prefixCommentIDs         = "biz#commentids#objID:%d:objType:%d:rootID:%d:sortType:%d"
	prefixCommentObjSortType = "biz#commentobj#sorttype#objID:%d:objType:%d:rootID:%d:sortType:%d"

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

	if in.ObjID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if in.ObjType <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjTypeRequired, "obj_type is required")
	}

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

	commentIDs, _ := l.cacheCommentIDs(in.ObjID, in.ObjType, in.RootID, cursor, pageSize, in.SortType)
	if len(commentIDs) > 0 {
		isCache = true
		if commentIDs[len(commentIDs)-1] == -1 {
			isEnd = true
		}

		comments, err = l.svcCtx.CommentModel.CacheCommentsByIDs(l.ctx, in.ObjID, commentIDs)
		if err != nil {
			return nil, err
		}

		commentMap := make(map[int64]*model.Comment, len(comments))
		for _, c := range comments {
			commentMap[c.ID] = c
		}
		for _, id := range commentIDs {
			if id <= 0 {
				continue
			}
			c, ok := commentMap[id]
			if !ok {
				continue
			}
			curPage = append(curPage, toCommentResponse(c))
		}

	} else {
		// 从数据库查询
		sfKey := fmt.Sprintf("commentsByObjID:%d:%d:%d:%d:%s", in.ObjID, in.ObjType, in.RootID, in.ReplyID, sortFiled)
		v, err, _ := l.svcCtx.SignleFlightGroup.Do(sfKey, func() (any, error) {
			return l.svcCtx.CommentModel.CommentListByObjID(l.ctx, in.ObjID, in.ObjType, in.RootID, in.ReplyID, sortFiled, types.DefaultLimit)
		})
		if err != nil {
			return nil, err
		}
		comments = v.([]*model.Comment)
		var firstPageComments []*model.Comment
		if len(comments) > int(pageSize) {
			firstPageComments = comments[:pageSize]
		} else {
			firstPageComments = comments
			isEnd = true
		}

		for _, c := range firstPageComments {
			curPage = append(curPage, toCommentResponse(c))
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
			err = l.addComments(context.Background(), in.ObjID, in.ObjType, in.RootID, in.SortType, comments)
			if err != nil {
				l.Logger.Errorf("addComments error: %v", err)
			}

		})
	}

	return ret, nil
}

func (l *GetCommentListLogic) cacheCommentIDs(objID int64, objType int64, rootID int64, cursor int64, pageSize int64, sortType int64) ([]int64, error) {
	if l.svcCtx == nil || l.svcCtx.BizRedis == nil {
		return nil, nil
	}

	key := formatCommentIDsKey(objID, objType, rootID, sortType)

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

func (l *GetCommentListLogic) addComments(ctx context.Context, objID int64, objType int64, rootID int64, sortType int64, comments []*model.Comment) error {
	if l.svcCtx == nil || l.svcCtx.BizRedis == nil {
		return nil
	}

	if len(comments) == 0 {
		return nil
	}
	keyPrimary := formatCommentIDsKey(objID, objType, rootID, sortType)
	keyCompat := formatCommentObjSortTypeKey(objID, objType, rootID, sortType)
	for _, c := range comments {
		var baseScore int64
		if sortType == types.SortCreatedTime {
			baseScore = c.CreatedAt.Unix()
		} else {
			baseScore = c.LikeCount
		}

		score := buildCompositeScore(c.Attrs, baseScore)
		if score < 0 {
			score = 0
		}

		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, keyPrimary, score, strconv.FormatInt(c.ID, 10))
		if err != nil {
			l.Logger.Errorf("addComments error: %v", err)
			return err
		}
		_, err = l.svcCtx.BizRedis.ZaddCtx(ctx, keyCompat, score, strconv.FormatInt(c.ID, 10))
		if err != nil {
			l.Logger.Errorf("addComments error: %v", err)
			return err
		}
	}

	if err := l.svcCtx.BizRedis.ExpireCtx(ctx, keyPrimary, commentIDsExpire); err != nil {
		return err
	}
	return l.svcCtx.BizRedis.ExpireCtx(ctx, keyCompat, commentIDsExpire)
}
