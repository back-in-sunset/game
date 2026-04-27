package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type UnSetCommentAttrsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnSetCommentAttrsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnSetCommentAttrsLogic {
	return &UnSetCommentAttrsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论取消置顶
func (l *UnSetCommentAttrsLogic) UnSetCommentAttrs(in *comment.UnSetCommentAttrsRequest) (*comment.UnSetCommentAttrsResponse, error) {
	if in.ObjID <= 0 {
		return nil, status.Error(400, "obj_id不能为空")
	}
	if in.CommentID <= 0 {
		return nil, status.Error(400, "comment_id不能为空")
	}

	c, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.CommentID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(404, "评论不存在")
		}
		return nil, status.Error(500, err.Error())
	}

	nextAttrs := c.Attrs &^ attrsPinnedBit
	c, err = l.svcCtx.CommentModel.SetCommentAttrs(l.ctx, in.ObjID, in.CommentID, nextAttrs)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(404, "评论不存在")
		}
		return nil, status.Error(500, err.Error())
	}
	if err = syncCommentScores(l.ctx, l.svcCtx, c); err != nil {
		return nil, status.Error(500, err.Error())
	}

	return &comment.UnSetCommentAttrsResponse{
		Success: true,
		Message: "ok",
	}, nil
}
