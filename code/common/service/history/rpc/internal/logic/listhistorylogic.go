package logic

import (
	"context"
	"net/http"

	"history/internal/errx"
	"history/model"
	"history/rpc/historyclient"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListHistoryLogic {
	return &ListHistoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListHistoryLogic) ListHistory(in *historyclient.ListHistoryRequest) (*historyclient.ListHistoryResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if in.MediaType != 0 {
		if err := validateMedia(in.MediaType, 1); err != nil {
			return nil, err
		}
	}
	if in.PageSize < 0 || in.PageSize > model.MaxPageSize {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodePageSizeInvalid, "page_size invalid")
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = model.DefaultPageSize
	}
	result, err := l.svcCtx.HistoryModel.List(l.ctx, in.UserID, in.MediaType, in.Cursor, in.LastID, pageSize)
	if err != nil {
		return nil, mapModelError(err)
	}
	out := &historyclient.ListHistoryResponse{
		List:   make([]*historyclient.HistoryRecord, 0, len(result.Records)),
		IsEnd:  result.IsEnd,
		Cursor: result.Cursor,
		LastID: result.LastID,
	}
	for _, r := range result.Records {
		out.List = append(out.List, toRPCRecord(r))
	}
	return out, nil
}
