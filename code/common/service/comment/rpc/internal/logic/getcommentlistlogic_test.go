package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"comment/rpc/types"

	"google.golang.org/grpc/status"
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

func TestGetCommentListLogic_InvalidParams(t *testing.T) {
	l := NewGetCommentListLogic(context.Background(), &svc.ServiceContext{
		CommentModel: &listCommentStubModel{},
	})

	_, err := l.GetCommentList(&comment.CommentListRequest{
		ObjID:    0,
		ObjType:  1,
		SortType: types.SortCreatedTime,
		PageSize: 10,
	})
	if err == nil {
		t.Fatalf("expected error for empty obj_id")
	}
	if got := int(status.Code(err)); got != 400 {
		t.Fatalf("status code=%d want=400", got)
	}

	_, err = l.GetCommentList(&comment.CommentListRequest{
		ObjID:    1001,
		ObjType:  0,
		SortType: types.SortCreatedTime,
		PageSize: 10,
	})
	if err == nil {
		t.Fatalf("expected error for empty obj_type")
	}
	if got := int(status.Code(err)); got != 400 {
		t.Fatalf("status code=%d want=400", got)
	}
}

func TestGetCommentListLogic_IsEndOnLastPage(t *testing.T) {
	now := time.Now()
	stub := &listCommentStubModel{
		listResp: []*model.Comment{
			{ID: 11, ObjID: 1001, ObjType: 1, MemberID: 3001, Message: "c1", CreatedAt: now},
			{ID: 12, ObjID: 1001, ObjType: 1, MemberID: 3002, Message: "c2", CreatedAt: now},
		},
	}
	l := NewGetCommentListLogic(context.Background(), &svc.ServiceContext{
		CommentModel: stub,
	})

	resp, err := l.GetCommentList(&comment.CommentListRequest{
		ObjID:    1001,
		ObjType:  1,
		SortType: types.SortCreatedTime,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetCommentList() error=%v", err)
	}
	if !resp.IsEnd {
		t.Fatalf("resp.IsEnd=false want=true")
	}
	if len(resp.Comments) != 2 {
		t.Fatalf("len(resp.Comments)=%d want=2", len(resp.Comments))
	}
}

func TestGetCommentListLogic_SortSwitch(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		sortType int64
		wantSort string
	}{
		{name: "sort by time", sortType: types.SortCreatedTime, wantSort: "created_at"},
		{name: "sort by like", sortType: types.SortLikeCount, wantSort: "like_count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &listCommentStubModel{
				listResp: []*model.Comment{
					{ID: 21, ObjID: 1002, ObjType: 1, MemberID: 3001, Message: "ok", CreatedAt: now},
				},
			}
			l := NewGetCommentListLogic(context.Background(), &svc.ServiceContext{
				CommentModel: stub,
			})

			_, err := l.GetCommentList(&comment.CommentListRequest{
				ObjID:    1002,
				ObjType:  1,
				SortType: tt.sortType,
				PageSize: 10,
			})
			if err != nil {
				t.Fatalf("GetCommentList() error=%v", err)
			}
			if stub.gotSort != tt.wantSort {
				t.Fatalf("got sort=%s want=%s", stub.gotSort, tt.wantSort)
			}
		})
	}
}
