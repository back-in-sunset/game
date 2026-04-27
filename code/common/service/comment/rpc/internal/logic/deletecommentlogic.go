package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除评论
func (l *DeleteCommentLogic) DeleteComment(in *comment.CommentRequest) (*comment.CommentResponse, error) {
	if in.ObjID <= 0 {
		return nil, status.Error(400, "obj_id不能为空")
	}
	if in.CommentID <= 0 {
		return nil, status.Error(400, "comment_id不能为空")
	}
	if in.MemberID <= 0 {
		return nil, status.Error(400, "member_id不能为空")
	}

	res, err := l.svcCtx.CommentModel.DeleteComment(l.ctx, in.ObjID, in.CommentID, in.MemberID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(404, "评论不存在")
		}
		return nil, status.Error(500, err.Error())
	}

	return toCommentResponse(res), nil
}
