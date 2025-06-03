package logic

import (
	"context"

	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnAuditCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnAuditCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnAuditCommentLogic {
	return &UnAuditCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论取消审核
func (l *UnAuditCommentLogic) UnAuditComment(in *comment.UnAuditCommentRequest) (*comment.UnAuditCommentResponse, error) {
	// todo: add your logic here and delete this line

	return &comment.UnAuditCommentResponse{}, nil
}
