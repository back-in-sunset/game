package logic

import (
	"context"
	"errors"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/eventbus"
	"comment/rpc/internal/svc"
	"comment/rpc/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnLikeCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnLikeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnLikeCommentLogic {
	return &UnLikeCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论取消点赞
func (l *UnLikeCommentLogic) UnLikeComment(in *comment.UnLikeCommentRequest) (*comment.UnLikeCommentResponse, error) {
	if in.ObjID <= 0 || in.ObjType <= 0 || in.CommentID <= 0 || in.MemberID <= 0 {
		return nil, errors.New("obj_id, obj_type, comment_id, member_id are required")
	}
	c, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.CommentID)
	if err != nil {
		return nil, err
	}

	res, err := execUnlikeScript(l.ctx, l.svcCtx, in.ObjID, in.ObjType, c.RootID, in.CommentID, in.MemberID)
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
			Action:    eventbus.LikeActionUnlike,
			ObjID:     in.ObjID,
			ObjType:   in.ObjType,
			CommentID: in.CommentID,
			MemberID:  in.MemberID,
			Delta:     -1,
			Ts:        time.Now().UnixMilli(),
		})
		if err != nil {
			return nil, err
		}
	}

	msg := "ok"
	if !res.Changed {
		msg = "already unliked"
	}

	return &comment.UnLikeCommentResponse{
		Success: true,
		Message: msg,
	}, nil
}
