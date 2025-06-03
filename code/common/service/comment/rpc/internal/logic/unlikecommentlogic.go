package logic

import (
	"context"

	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

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
	// todo: add your logic here and delete this line

	return &comment.UnLikeCommentResponse{}, nil
}
