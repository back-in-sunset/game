package logic

import (
	"context"
	"net/http"
	"strings"

	"comment/api/commentclient"
	"comment/api/internal/svc"
	"comment/api/internal/types"
	"comment/internal/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddLogic) Add(req *types.CommentRequest) (resp *types.CommentResponse, err error) {
	req.Message = strings.TrimSpace(req.Message)
	if req.ObjID <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeObjIDRequired, "obj_id is required")
	}
	if req.ObjType <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeObjTypeRequired, "obj_type is required")
	}
	if req.MemberID <= 0 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeMemberIDRequired, "member_id is required")
	}
	if req.Message == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeMessageRequired, "message is required")
	}

	res, err := l.svcCtx.CommentRpc.AddComment(l.ctx, &commentclient.CommentRequest{
		ObjID:       req.ObjID,
		ObjType:     req.ObjType,
		MemberID:    req.MemberID,
		CommentID:   req.CommentID,
		AtMemberIDs: req.AtMemberIDs,
		Ip:          req.Ip,
		Platform:    req.Platform,
		Device:      req.Device,
		Message:     req.Message,
		Meta:        req.Meta,
		ReplyID:     req.ReplyID,
		State:       req.State,
		RootID:      req.RootID,
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
