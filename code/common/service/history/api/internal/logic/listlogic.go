package logic

import (
	"context"

	"history/api/internal/svc"
	"history/api/internal/types"
	"history/rpc/historyclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListLogic) List(req *types.ListHistoryRequest) (*types.ListHistoryResponse, error) {
	uid, err := extractUserID(l.ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.HistoryRpc.ListHistory(l.ctx, &historyclient.ListHistoryRequest{
		UserID:    uid,
		MediaType: req.MediaType,
		Cursor:    req.Cursor,
		LastID:    req.LastID,
		PageSize:  req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := &types.ListHistoryResponse{
		List:   make([]types.HistoryRecord, 0, len(res.List)),
		IsEnd:  res.IsEnd,
		Cursor: res.Cursor,
		LastID: res.LastID,
	}
	for _, item := range res.List {
		if converted := toAPIRecord(item); converted != nil {
			out.List = append(out.List, *converted)
		}
	}
	return out, nil
}
