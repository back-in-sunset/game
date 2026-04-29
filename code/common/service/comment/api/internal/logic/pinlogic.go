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

type PinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PinLogic {
	return &PinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PinLogic) Pin(req *types.CommentActionRequest) (*types.CommentActionResponse, error) {
	if req.ObjID <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if req.ObjType <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeObjTypeRequired, "obj_type is required")
	}
	if req.CommentID <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeCommentIDRequired, "comment_id is required")
	}

	res, err := l.svcCtx.CommentRpc.SetCommentAttrs(l.ctx, &commentclient.SetCommentAttrsRequest{
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
