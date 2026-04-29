package logic

import (
	"context"

	"history/api/internal/svc"
	"history/api/internal/types"
	"history/rpc/historyclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordLogic {
	return &RecordLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RecordLogic) Record(req *types.RecordHistoryRequest) (*types.RecordHistoryResponse, error) {
	uid, err := extractUserID(l.ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.HistoryRpc.RecordHistory(l.ctx, &historyclient.RecordHistoryRequest{
		UserID:     uid,
		MediaType:  req.MediaType,
		MediaID:    req.MediaID,
		Title:      req.Title,
		Cover:      req.Cover,
		AuthorID:   req.AuthorID,
		ProgressMs: req.ProgressMs,
		DurationMs: req.DurationMs,
		Finished:   req.Finished,
		Source:     req.Source,
		Device:     req.Device,
		Meta:       req.Meta,
	})
	if err != nil {
		return nil, err
	}
	return &types.RecordHistoryResponse{Record: toAPIRecord(res.Record)}, nil
}
