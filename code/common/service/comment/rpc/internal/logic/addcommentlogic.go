package logic

import (
	"context"

	"comment/model"
	"comment/rpc/comment"
	"comment/rpc/internal/svc"

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
	l.svcCtx.CommentModel.AddComment(l.ctx,
		&model.CommentSubject{
			ObjID:    in.ObjID,
			ObjType:  in.ObjType,
			MemberID: in.MemberID,
			State:    in.State,
		},
		&model.CommentIndex{
			ObjID:    in.ObjID,
			ObjType:  in.ObjType,
			MemberID: in.MemberID,
			RootID:   in.RootID,
			ReplyID:  in.ReplyID,
			// Floor:     0,
			// Count:     0,
			// RootCount: 0,
			// LikeCount: 0,
			// HateCount: 0,
			State: in.State,
			// Attrs:     0,
			// CreatedAt: time.Time{},
			// UpdatedAt: time.Time{},
		},
		&model.CommentContent{
			// CommentID:   0,
			ObjID:       in.ObjID,
			AtMemberIDs: in.AtMemberIDs,
			Ip:          in.Ip,
			Platform:    in.Platform,
			Device:      in.Device,
			Message:     in.Message,
			Meta:        in.Meta,
			// CreatedAt:   time.Time{},
			// UpdatedAt:   time.Time{},
		},
	)

	return &comment.CommentResponse{}, nil
}
