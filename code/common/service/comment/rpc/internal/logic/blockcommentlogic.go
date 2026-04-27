package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlockCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlockCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlockCommentLogic {
	return &BlockCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论屏蔽
func (l *BlockCommentLogic) BlockComment(in *comment.BlockCommentRequest) (*comment.BlockCommentResponse, error) {
	// todo: add your logic here and delete this line

	return &comment.BlockCommentResponse{}, nil
}
