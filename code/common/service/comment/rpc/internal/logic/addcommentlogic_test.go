package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
)

type addCommentStubModel struct {
	addResp  *model.CommentSchema
	findResp *model.Comment
	addErr   error
	findErr  error

	gotSubject *model.CommentSubject
	gotIndex   *model.CommentIndex
	gotContent *model.CommentContent
}

func (m *addCommentStubModel) AddComment(ctx context.Context, data *model.CommentSubject, ci *model.CommentIndex, cc *model.CommentContent) (*model.CommentSchema, error) {
	m.gotSubject = data
	m.gotIndex = ci
	m.gotContent = cc
	return m.addResp, m.addErr
}

func (m *addCommentStubModel) DeleteComment(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *addCommentStubModel) CommentListByObjID(context.Context, int64, int64, int64, int64, string, int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *addCommentStubModel) FindOneByObjID(context.Context, int64, int64) (*model.Comment, error) {
	return m.findResp, m.findErr
}

func (m *addCommentStubModel) CacheCommentsByIDs(context.Context, int64, []int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *addCommentStubModel) AdjustCommentLikeCount(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (m *addCommentStubModel) SetCommentState(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *addCommentStubModel) SetCommentAttrs(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func TestAddCommentLogic_LevelCases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		req       *comment.CommentRequest
		findResp  *model.Comment
		wantRoot  int64
		wantReply int64
	}{
		{
			name: "first level comment",
			req: &comment.CommentRequest{
				ObjID:    1001,
				ObjType:  1,
				MemberID: 2001,
				Message:  " 一级评论 ",
				RootID:   0,
				ReplyID:  0,
			},
			findResp: &model.Comment{
				ID:        9001,
				ObjID:     1001,
				ObjType:   1,
				MemberID:  2001,
				RootID:    0,
				ReplyID:   0,
				Message:   "一级评论",
				CreatedAt: now,
			},
			wantRoot:  0,
			wantReply: 0,
		},
		{
			name: "second level comment",
			req: &comment.CommentRequest{
				ObjID:    1001,
				ObjType:  1,
				MemberID: 2002,
				Message:  " 二级评论 ",
				RootID:   9001,
				ReplyID:  9001,
			},
			findResp: &model.Comment{
				ID:        9002,
				ObjID:     1001,
				ObjType:   1,
				MemberID:  2002,
				RootID:    9001,
				ReplyID:   9001,
				Message:   "二级评论",
				CreatedAt: now,
			},
			wantRoot:  9001,
			wantReply: 9001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &addCommentStubModel{
				addResp:  &model.CommentSchema{CommentID: tt.findResp.ID},
				findResp: tt.findResp,
			}
			l := NewAddCommentLogic(context.Background(), &svc.ServiceContext{
				CommentModel: stub,
			})

			resp, err := l.AddComment(tt.req)
			if err != nil {
				t.Fatalf("AddComment() error = %v", err)
			}
			if stub.gotIndex == nil {
				t.Fatalf("gotIndex is nil")
			}
			if stub.gotIndex.RootID != tt.wantRoot {
				t.Fatalf("got RootID=%d want=%d", stub.gotIndex.RootID, tt.wantRoot)
			}
			if stub.gotIndex.ReplyID != tt.wantReply {
				t.Fatalf("got ReplyID=%d want=%d", stub.gotIndex.ReplyID, tt.wantReply)
			}
			if stub.gotContent == nil || stub.gotContent.Message == "" {
				t.Fatalf("message not captured")
			}
			if resp.RootID != tt.wantRoot {
				t.Fatalf("resp RootID=%d want=%d", resp.RootID, tt.wantRoot)
			}
			if resp.ReplyID != tt.wantReply {
				t.Fatalf("resp ReplyID=%d want=%d", resp.ReplyID, tt.wantReply)
			}
		})
	}
}
