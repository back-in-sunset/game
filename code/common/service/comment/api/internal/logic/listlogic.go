package logic

import (
	"context"

	"comment/api/internal/svc"
	"comment/api/internal/types"
	"comment/rpc/comment"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/singleflight"
)

type ListLogic struct {
	logx.Logger
	ctx          context.Context
	svcCtx       *svc.ServiceContext
	singleflight singleflight.Group
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List retrieves a list of comments based on the request parameters.
func (l *ListLogic) List(req *types.CommentListRequest) (resp *types.CommentListResponse, err error) {
	var commentreq comment.CommentListRequest
	copier.Copy(&commentreq, req)

	commentListResp, err := l.svcCtx.CommentRpc.GetCommentList(l.ctx, &commentreq)
	if err != nil {
		return nil, err
	}
	resp = &types.CommentListResponse{
		List: make([]types.CommentResponse, 0, 100),
	}
	copier.Copy(&resp.List, commentListResp.Comments)
	return
}
