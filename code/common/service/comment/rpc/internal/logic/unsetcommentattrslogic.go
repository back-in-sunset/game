package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
	// todo: add your logic here and delete this line

	return &comment.UnSetCommentAttrsResponse{}, nil
}
