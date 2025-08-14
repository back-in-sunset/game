package logic

import (
	"context"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
	return &GetCommentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取评论列表
func (l *GetCommentListLogic) GetCommentList(in *comment.CommentListRequest) (*comment.CommentListResponse, error) {
	comments, err := l.svcCtx.CommentModel.CommentByObjID(l.ctx, in.ObjID, in.ObjType, "like_count",
		in.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var commentResponses []*comment.CommentResponse
	copier.Copy(&commentResponses, comments)

	return &comment.CommentListResponse{
		Comments: commentResponses,
		IsEnd:    false,
		Cursor:   0,
		LastID:   0,
	}, nil
}
