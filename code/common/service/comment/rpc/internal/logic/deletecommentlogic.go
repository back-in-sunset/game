package logic

import (
	"context"
	"net/http"

	"comment/internal/errx"
	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除评论
func (l *DeleteCommentLogic) DeleteComment(in *comment.CommentRequest) (*comment.CommentResponse, error) {
	if in.ObjID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if in.CommentID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeCommentIDRequired, "comment_id is required")
	}
	if in.MemberID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeMemberIDRequired, "member_id is required")
	}

	res, err := l.svcCtx.CommentModel.DeleteComment(l.ctx, in.ObjID, in.CommentID, in.MemberID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errx.RPCError(http.StatusNotFound, errx.CodeCommentNotFound, "comment not found")
		}
		if err == model.ErrPermissionDenied {
			return nil, errx.RPCError(http.StatusForbidden, errx.CodePermissionDenied, "permission denied")
		}
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}
	if err = removeCommentScores(l.ctx, l.svcCtx, res.ObjID, res.ObjType, res.RootID, res.ID); err != nil {
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	return toCommentResponse(res), nil
}
