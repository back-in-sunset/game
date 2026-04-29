package logic

import (
	"context"
	"net/http"

	"comment/api/commentclient"
	"comment/api/internal/svc"
	"comment/api/internal/types"
	"comment/internal/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnblockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnblockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblockLogic {
	return &UnblockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnblockLogic) Unblock(req *types.CommentActionRequest) (*types.CommentActionResponse, error) {
	if req.ObjID <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if req.ObjType <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeObjTypeRequired, "obj_type is required")
	}
	if req.CommentID <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeCommentIDRequired, "comment_id is required")
	}

	res, err := l.svcCtx.CommentRpc.UnBlockComment(l.ctx, &commentclient.UnBlockCommentRequest{
		ObjID:     req.ObjID,
		ObjType:   req.ObjType,
		CommentID: req.CommentID,
		MemberID:  req.MemberID,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommentActionResponse{
		Success: res.Success,
		Message: res.Message,
	}, nil
}
