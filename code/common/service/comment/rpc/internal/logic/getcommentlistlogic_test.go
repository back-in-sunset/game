package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"comment/rpc/types"
)

type listCommentStubModel struct {
	gotObjID   int64
	gotObjType int64
	gotRootID  int64
	gotReplyID int64
	gotSort    string
	gotLimit   int64
	listResp   []*model.Comment
	listErr    error
}

func (m *listCommentStubModel) AddComment(context.Context, *model.CommentSubject, *model.CommentIndex, *model.CommentContent) (*model.CommentSchema, error) {
	return nil, nil
}

func (m *listCommentStubModel) DeleteComment(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *listCommentStubModel) CommentListByObjID(_ context.Context, objID int64, objType int64, rootID int64, replyID int64, sortField string, limit int64) ([]*model.Comment, error) {
	m.gotObjID = objID
	m.gotObjType = objType
	m.gotRootID = rootID
	m.gotReplyID = replyID
	m.gotSort = sortField
	m.gotLimit = limit
	return m.listResp, m.listErr
}

func (m *listCommentStubModel) FindOneByObjID(context.Context, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *listCommentStubModel) CacheCommentsByIDs(context.Context, int64, []int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *listCommentStubModel) AdjustCommentLikeCount(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (m *listCommentStubModel) SetCommentState(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *listCommentStubModel) SetCommentAttrs(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func TestGetCommentListLogic_LevelLists(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		req         *comment.CommentListRequest
		wantRootID  int64
		wantReplyID int64
	}{
		{
			name: "first-level list",
			req: &comment.CommentListRequest{
				ObjID:    1001,
				ObjType:  1,
				RootID:   0,
				SortType: types.SortCreatedTime,
				PageSize: 10,
			},
			wantRootID:  0,
			wantReplyID: 0,
		},
		{
			name: "second-level list by root",
			req: &comment.CommentListRequest{
				ObjID:    1001,
				ObjType:  1,
				RootID:   9001,
				SortType: types.SortCreatedTime,
				PageSize: 10,
			},
			wantRootID:  9001,
			wantReplyID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &listCommentStubModel{
				listResp: []*model.Comment{
					{
						ID:        1,
						ObjID:     tt.req.ObjID,
						ObjType:   tt.req.ObjType,
						RootID:    tt.wantRootID,
						ReplyID:   tt.wantReplyID,
						MemberID:  3001,
						Message:   "ok",
						CreatedAt: now,
					},
				},
			}
			l := NewGetCommentListLogic(context.Background(), &svc.ServiceContext{
				CommentModel: stub,
			})

			resp, err := l.GetCommentList(tt.req)
			if err != nil {
				t.Fatalf("GetCommentList() error=%v", err)
			}
			if stub.gotRootID != tt.wantRootID {
				t.Fatalf("got rootID=%d want=%d", stub.gotRootID, tt.wantRootID)
			}
			if stub.gotReplyID != tt.wantReplyID {
				t.Fatalf("got replyID=%d want=%d", stub.gotReplyID, tt.wantReplyID)
			}
			if stub.gotSort != "created_at" {
				t.Fatalf("got sort=%s want=created_at", stub.gotSort)
			}
			if len(resp.Comments) != 1 {
				t.Fatalf("len(resp.Comments)=%d want=1", len(resp.Comments))
			}
			if resp.Comments[0].RootID != tt.wantRootID {
				t.Fatalf("resp rootID=%d want=%d", resp.Comments[0].RootID, tt.wantRootID)
			}
		})
	}
}
