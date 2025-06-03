package logic

import (
	"context"

	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetCommentAttrsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetCommentAttrsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetCommentAttrsLogic {
	return &SetCommentAttrsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论置顶
func (l *SetCommentAttrsLogic) SetCommentAttrs(in *comment.SetCommentAttrsRequest) (*comment.SetCommentAttrsResponse, error) {
	// todo: add your logic here and delete this line

	return &comment.SetCommentAttrsResponse{}, nil
}
