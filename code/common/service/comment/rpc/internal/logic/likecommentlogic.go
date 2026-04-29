package logic

import (
	"context"
	"net/http"
	"time"

	"comment/internal/errx"
	"comment/rpc/comment"
	"comment/rpc/internal/eventbus"
	"comment/rpc/internal/svc"
	"comment/rpc/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeCommentLogic {
	return &LikeCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论点赞
func (l *LikeCommentLogic) LikeComment(in *comment.LikeCommentRequest) (*comment.LikeCommentResponse, error) {
	if in.ObjID <= 0 || in.ObjType <= 0 || in.CommentID <= 0 || in.MemberID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeBadRequestDefault, "obj_id, obj_type, comment_id, member_id are required")
	}
	c, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.CommentID)
	if err != nil {
		return nil, err
	}

	res, err := execLikeScript(l.ctx, l.svcCtx, in.ObjID, in.ObjType, c.RootID, in.CommentID, in.MemberID)
	if err != nil {
		return nil, err
	}
	if res.Changed {
		likeScore := buildCompositeScore(c.Attrs, res.LikeCount)
		timeScore := buildCompositeScore(c.Attrs, c.CreatedAt.Unix())
		if err = updateCommentScoresForSort(l.ctx, l.svcCtx, in.ObjID, in.ObjType, c.RootID, in.CommentID, types.SortLikeCount, likeScore); err != nil {
			return nil, err
		}
		if err = updateCommentScoresForSort(l.ctx, l.svcCtx, in.ObjID, in.ObjType, c.RootID, in.CommentID, types.SortCreatedTime, timeScore); err != nil {
			return nil, err
		}

		err = l.svcCtx.LikeEventBus.PublishLikeEvent(l.ctx, eventbus.LikeEvent{
			Action:    eventbus.LikeActionLike,
			ObjID:     in.ObjID,
			ObjType:   in.ObjType,
			CommentID: in.CommentID,
			MemberID:  in.MemberID,
			Delta:     1,
			Ts:        time.Now().UnixMilli(),
		})
		if err != nil {
			return nil, err
		}
	}

	msg := "ok"
	if !res.Changed {
		msg = "already liked"
	}

	return &comment.LikeCommentResponse{
		Success: true,
		Message: msg,
	}, nil
}
