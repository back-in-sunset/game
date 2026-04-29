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

type BlockCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlockCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlockCommentLogic {
	return &BlockCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论屏蔽
func (l *BlockCommentLogic) BlockComment(in *comment.BlockCommentRequest) (*comment.BlockCommentResponse, error) {
	if in.ObjID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if in.CommentID <= 0 {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodeCommentIDRequired, "comment_id is required")
	}

	c, err := l.svcCtx.CommentModel.SetCommentState(l.ctx, in.ObjID, in.CommentID, 1)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errx.RPCError(http.StatusNotFound, errx.CodeCommentNotFound, "comment not found")
		}
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}
	if err = syncCommentScores(l.ctx, l.svcCtx, c); err != nil {
		return nil, errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
	}

	return &comment.BlockCommentResponse{
		Success: true,
		Message: "ok",
	}, nil
}
