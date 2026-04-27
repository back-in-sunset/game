package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type GetCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentLogic {
	return &GetCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取评论
func (l *GetCommentLogic) GetComment(in *comment.CommentRequest) (*comment.CommentResponse, error) {
	if in.ObjID <= 0 {
		return nil, status.Error(400, "obj_id不能为空")
	}
	if in.CommentID <= 0 {
		return nil, status.Error(400, "comment_id不能为空")
	}

	res, err := l.svcCtx.CommentModel.FindOneByObjID(l.ctx, in.ObjID, in.CommentID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(404, "评论不存在")
		}
		return nil, status.Error(500, err.Error())
	}

	return toCommentResponse(res), nil
}
