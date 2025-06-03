package logic

import (
	"context"

	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportCommentLogic {
	return &ReportCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论举报
func (l *ReportCommentLogic) ReportComment(in *comment.ReportCommentRequest) (*comment.ReportCommentResponse, error) {
	// todo: add your logic here and delete this line

	return &comment.ReportCommentResponse{}, nil
}
