package logic

import (
	"context"

	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuditCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditCommentLogic {
	return &AuditCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通过评论审核
func (l *AuditCommentLogic) AuditComment(in *comment.AuditCommentRequest) (*comment.AuditCommentResponse, error) {
	// todo: add your logic here and delete this line

	return &comment.AuditCommentResponse{}, nil
}
