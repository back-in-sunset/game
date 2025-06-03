package logic

import (
	"context"

	"comment/model"
	"comment/rpc/internal/svc"
	"comment/rpc/pb/comment"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
	return &AddCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加评论
func (l *AddCommentLogic) AddComment(in *comment.CommentRequest) (*comment.CommentResponse, error) {
	// todo: add your logic here and delete this line
	l.svcCtx.CommentSubjectModel.Insert(l.ctx, &model.CommentSubject{})
	// 这里可以添加更多的逻辑，比如发送通知等
	// 返回评论响应
	return &comment.CommentResponse{}, nil
}
