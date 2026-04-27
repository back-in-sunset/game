package logic

import (
	"context"
	"errors"

	"comment/api/commentclient"
	"comment/api/internal/svc"
	"comment/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteLogic {
	return &DeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteLogic) Delete(req *types.CommentRequest) (resp *types.CommentResponse, err error) {
	if req.ObjID <= 0 {
		return nil, errors.New("obj_id is required")
	}
	if req.CommentID <= 0 {
		return nil, errors.New("comment_id is required")
	}
	if req.MemberID <= 0 {
		return nil, errors.New("member_id is required")
	}

	res, err := l.svcCtx.CommentRpc.DeleteComment(l.ctx, &commentclient.CommentRequest{
		ObjID:     req.ObjID,
		ObjType:   req.ObjType,
		MemberID:  req.MemberID,
		CommentID: req.CommentID,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommentResponse{
		ID:        res.ID,
		ObjID:     res.ObjID,
		ObjType:   res.ObjType,
		MemberID:  res.MemberID,
		CommentID: res.CommentID,
		State:     res.State,
		ReplyID:   res.ReplyID,
		RootID:    res.RootID,
		CreatedAt: res.CreatedAt,
		Floor:     res.Floor,
		LikeCount: res.LikeCount,
		HateCount: res.HateCount,
		Count:     res.Count,
	}, nil
}
