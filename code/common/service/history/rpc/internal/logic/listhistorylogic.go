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
	records, err := l.svcCtx.HistoryModel.ListByUser(l.ctx, in.UserID, in.MediaType, in.Cursor, in.LastID, pageSize+1)
	if err != nil {
		return nil, mapModelError(err)
	}
	isEnd := true
	if len(records) > int(pageSize) {
		isEnd = false
		records = records[:pageSize]
	}
	out := &historyclient.ListHistoryResponse{List: make([]*historyclient.HistoryRecord, 0, len(records)), IsEnd: isEnd}
	for _, r := range records {
		out.List = append(out.List, toRPCRecord(r))
	}
	if len(out.List) > 0 {
		last := out.List[len(out.List)-1]
		out.Cursor = last.LastSeenAt
		out.LastID = last.ID
	}
	return out, nil
}
