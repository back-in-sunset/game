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
	if in.ObjID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if in.CommentID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeCommentIDRequired, "comment_id is required")
	}

	c, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.CommentID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errx.RPCError(http.StatusNotFound, errx.CodeCommentNotFound, "comment not found")
		}
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	nextAttrs := c.Attrs | attrsPinnedBit
	c, err = l.svcCtx.CommentModel.SetCommentAttrs(l.ctx, in.ObjID, in.CommentID, nextAttrs)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errx.RPCError(http.StatusNotFound, errx.CodeCommentNotFound, "comment not found")
		}
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}
	if err = syncCommentScores(l.ctx, l.svcCtx, c); err != nil {
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	return &comment.SetCommentAttrsResponse{
		Success: true,
		Message: "ok",
	}, nil
}
