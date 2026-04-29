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

type GetCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentLogic {
	return &GetCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取评论
func (l *GetCommentLogic) GetComment(in *comment.CommentRequest) (*comment.CommentResponse, error) {
	if in.ObjID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if in.CommentID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeCommentIDRequired, "comment_id is required")
	}

	res, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.CommentID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errx.RPCError(http.StatusNotFound, errx.CodeCommentNotFound, "comment not found")
		}
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	return toCommentResponse(res), nil
}
