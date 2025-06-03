package logic

import (
	"context"

	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
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
	// todo: add your logic here and delete this line

	return &comment.UnBlockCommentResponse{}, nil
}
