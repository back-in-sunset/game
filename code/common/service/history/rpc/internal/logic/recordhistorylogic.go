package logic

import (
	"context"

	"history/model"
	"history/rpc/historyclient"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordHistoryLogic {
	return &RecordHistoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RecordHistoryLogic) RecordHistory(in *historyclient.RecordHistoryRequest) (*historyclient.RecordHistoryResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := validateMedia(in.MediaType, in.MediaID); err != nil {
		return nil, err
	}
	finished := int64(0)
	if in.Finished {
		finished = 1
	}
	record, err := l.svcCtx.HistoryModel.UpsertRecord(l.ctx, &model.HistoryRecord{
		UserID:     in.UserID,
		MediaType:  in.MediaType,
		MediaID:    in.MediaID,
		Title:      in.Title,
		Cover:      in.Cover,
		AuthorID:   in.AuthorID,
		ProgressMs: in.ProgressMs,
		DurationMs: in.DurationMs,
		Finished:   finished,
		Source:     in.Source,
		Device:     in.Device,
		Meta:       in.Meta,
	})
	if err != nil {
		return nil, mapModelError(err)
	}
	return &historyclient.RecordHistoryResponse{Record: toRPCRecord(record)}, nil
}
