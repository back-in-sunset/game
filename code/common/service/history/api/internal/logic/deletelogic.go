package logic

import (
	"context"

	"history/api/internal/svc"
	"history/api/internal/types"
	"history/rpc/historyclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteLogic) DeleteItem(req *types.DeleteHistoryItemRequest) (*types.ActionResponse, error) {
	uid, err := extractUserID(l.ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.HistoryRpc.DeleteHistoryItem(l.ctx, &historyclient.DeleteHistoryItemRequest{UserID: uid, MediaType: req.MediaType, MediaID: req.MediaID})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Success: res.Success, Message: res.Message}, nil
}

func (l *DeleteLogic) ClearByType(req *types.ClearHistoryByTypeRequest) (*types.ActionResponse, error) {
	uid, err := extractUserID(l.ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.HistoryRpc.ClearHistoryByType(l.ctx, &historyclient.ClearHistoryByTypeRequest{UserID: uid, MediaType: req.MediaType})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Success: res.Success, Message: res.Message}, nil
}

func (l *DeleteLogic) ClearAll(req *types.ClearHistoryAllRequest) (*types.ActionResponse, error) {
	uid, err := extractUserID(l.ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.HistoryRpc.ClearHistoryAll(l.ctx, &historyclient.ClearHistoryAllRequest{UserID: uid})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Success: res.Success, Message: res.Message}, nil
}
