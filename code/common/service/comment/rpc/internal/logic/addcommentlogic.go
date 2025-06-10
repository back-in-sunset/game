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
	return l.svcCtx.CommentModel.AddComment(l.ctx, &model.CommentSubject{
		ObjId:    in.ObjId,
		ObjType:  in.ObjType,
		MemberId: in.MemberId,
		State:    0,
		Attrs:    0,
	}, &model.CommentIndex{
		ObjId:    in.ObjId,
		ObjType:  in.ObjType,
		MemberId: in.MemberId,
		RootId:   in.RootId,
		ReplyId:  in.ReplyId,
		State:    0,
		Attrs:    0,
	}, &model.CommentContent{
		ObjId:       in.ObjId,
		AtMemberIds: in.AtMemberIds,
		Ip:          in.Ip,
		Platform:    in.Platform,
		Device:      in.Device,
		Message:     in.Message,
		Meta:        in.Meta,
	})
}
