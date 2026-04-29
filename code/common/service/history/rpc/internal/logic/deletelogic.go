package logic

import (
	"context"

	"history/rpc/historyclient"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteHistoryLogic {
	return &DeleteHistoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteHistoryLogic) DeleteHistoryItem(in *historyclient.DeleteHistoryItemRequest) (*historyclient.ActionResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := validateMedia(in.MediaType, in.MediaID); err != nil {
		return nil, err
	}
	if err := l.svcCtx.HistoryModel.SoftDeleteItem(l.ctx, in.UserID, in.MediaType, in.MediaID); err != nil {
		return nil, mapModelError(err)
	}
	return &historyclient.ActionResponse{Success: true, Message: "ok"}, nil
}

func (l *DeleteHistoryLogic) ClearHistoryByType(in *historyclient.ClearHistoryByTypeRequest) (*historyclient.ActionResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := validateMedia(in.MediaType, 1); err != nil {
		return nil, err
	}
	if err := l.svcCtx.HistoryModel.SoftDeleteByType(l.ctx, in.UserID, in.MediaType); err != nil {
		return nil, mapModelError(err)
	}
	return &historyclient.ActionResponse{Success: true, Message: "ok"}, nil
}

func (l *DeleteHistoryLogic) ClearHistoryAll(in *historyclient.ClearHistoryAllRequest) (*historyclient.ActionResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := l.svcCtx.HistoryModel.SoftDeleteAll(l.ctx, in.UserID); err != nil {
		return nil, mapModelError(err)
	}
	return &historyclient.ActionResponse{Success: true, Message: "ok"}, nil
}
