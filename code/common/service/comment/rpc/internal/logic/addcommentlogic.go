package logic

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"comment/internal/errx"
	"comment/rpc/comment"
	"comment/rpc/internal/notify"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"comment/rpc/types"

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
	in.Message = strings.TrimSpace(in.Message)
	if in.ObjID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if in.ObjType <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjTypeRequired, "obj_type is required")
	}
	if in.MemberID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeMemberIDRequired, "member_id is required")
	}
	if in.Message == "" {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeMessageRequired, "message is required")
	}
	if len([]rune(in.Message)) > types.MaxCommentLength {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeMessageTooLong, "message length must be <= 1000")
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
		if errors.Is(err, model.ErrInvalidReply) {
			return nil, errx.RPCError(http.StatusBadRequest, errx.CodeInvalidReply, "invalid reply relation")
		}
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	commentData, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, res.CommentID)
	if err != nil {
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}
	if err = syncCommentScores(l.ctx, l.svcCtx, commentData); err != nil {
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}
	if err = l.notifyReplyIfNeeded(in, commentData); err != nil {
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	return toCommentResponse(commentData), nil
}

func (l *AddCommentLogic) notifyReplyIfNeeded(in *comment.CommentRequest, commentData *model.Comment) error {
	if in.ReplyID <= 0 || commentData == nil || l.svcCtx.CommentNotifier == nil {
		return nil
	}
	replyTarget, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.ReplyID)
	if err != nil {
		return err
	}
	if replyTarget == nil || replyTarget.MemberID <= 0 || replyTarget.MemberID == in.MemberID {
		return nil
	}
	scope := l.svcCtx.Config.ReplyNoticeScope
	domain := scope.Domain
	if domain == "" {
		domain = "platform"
	}
	return l.svcCtx.CommentNotifier.NotifyReply(l.ctx, notify.ReplyNotice{
		Domain:      domain,
		TenantID:    scope.TenantID,
		ProjectID:   scope.ProjectID,
		Environment: scope.Environment,
		ObjID:       in.ObjID,
		ObjType:     in.ObjType,
		CommentID:   commentData.ID,
		ReplyID:     in.ReplyID,
		SenderID:    in.MemberID,
		ReceiverID:  replyTarget.MemberID,
		Message:     commentData.Message,
	})
}
