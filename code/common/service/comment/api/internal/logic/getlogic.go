package logic

import (
	"context"
	"errors"

	"comment/api/commentclient"
	"comment/api/internal/svc"
	"comment/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLogic) Get(req *types.CommentRequest) (resp *types.CommentResponse, err error) {
	if req.ObjID <= 0 {
		return nil, errors.New("obj_id is required")
	}
	if req.CommentID <= 0 {
		return nil, errors.New("comment_id is required")
	}

	res, err := l.svcCtx.CommentRpc.GetComment(l.ctx, &commentclient.CommentRequest{
		ObjID:     req.ObjID,
		ObjType:   req.ObjType,
		CommentID: req.CommentID,
	})
	if err != nil {
		return nil, err
	}

	return &types.CommentResponse{
		ID:          res.ID,
		ObjID:       res.ObjID,
		ObjType:     res.ObjType,
		MemberID:    res.MemberID,
		CommentID:   res.CommentID,
		AtMemberIDs: res.AtMemberIDs,
		Ip:          res.Ip,
		Platform:    res.Platform,
		Device:      res.Device,
		Message:     res.Message,
		Meta:        res.Meta,
		ReplyID:     res.ReplyID,
		State:       res.State,
		RootID:      res.RootID,
		CreatedAt:   res.CreatedAt,
		Floor:       res.Floor,
		LikeCount:   res.LikeCount,
		HateCount:   res.HateCount,
		Count:       res.Count,
	}, nil
}
