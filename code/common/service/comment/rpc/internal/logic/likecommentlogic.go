package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"

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
	// todo: add your logic here and delete this line

	return &comment.LikeCommentResponse{}, nil
}
