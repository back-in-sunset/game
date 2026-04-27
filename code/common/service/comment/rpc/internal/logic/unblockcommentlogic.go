package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type UnBlockCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnBlockCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnBlockCommentLogic {
	return &UnBlockCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论取消屏蔽
func (l *UnBlockCommentLogic) UnBlockComment(in *comment.UnBlockCommentRequest) (*comment.UnBlockCommentResponse, error) {
	if in.ObjID <= 0 {
		return nil, status.Error(400, "obj_id不能为空")
	}
	if in.CommentID <= 0 {
		return nil, status.Error(400, "comment_id不能为空")
	}

	c, err := l.svcCtx.CommentModel.SetCommentState(l.ctx, in.ObjID, in.CommentID, 0)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(404, "评论不存在")
		}
		return nil, status.Error(500, err.Error())
	}
	if err = syncCommentScores(l.ctx, l.svcCtx, c); err != nil {
		return nil, status.Error(500, err.Error())
	}

	return &comment.UnBlockCommentResponse{
		Success: true,
		Message: "ok",
	}, nil
}
