package logic

import (
	"context"
	"strings"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
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
	in.Message = strings.TrimSpace(in.Message)
	if in.ObjID <= 0 {
		return nil, status.Error(400, "obj_id不能为空")
	}
	if in.ObjType <= 0 {
		return nil, status.Error(400, "obj_type不能为空")
	}
	if in.MemberID <= 0 {
		return nil, status.Error(400, "member_id不能为空")
	}
	if in.Message == "" {
		return nil, status.Error(400, "message不能为空")
	}

	res, err := l.svcCtx.CommentModel.AddComment(l.ctx,
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
	if err != nil {
		return nil, status.Error(500, err.Error())
	}

	commentData, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, res.CommentID)
	if err != nil {
		return nil, status.Error(500, err.Error())
	}
	if err = syncCommentScores(l.ctx, l.svcCtx, commentData); err != nil {
		return nil, status.Error(500, err.Error())
	}

	return toCommentResponse(commentData), nil
}
