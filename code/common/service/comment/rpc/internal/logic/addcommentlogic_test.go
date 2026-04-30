package logic

import (
	"context"
	"strings"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/config"
	"comment/rpc/internal/notify"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
	"comment/rpc/types"

	"google.golang.org/grpc/status"
)

type addCommentStubModel struct {
	addResp  *model.CommentSchema
	findResp map[int64]*model.Comment
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

func (m *addCommentStubModel) FindOneByObjID(_ context.Context, _ int64, id int64) (*model.Comment, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.findResp == nil {
		return nil, nil
	}
	return m.findResp[id], nil
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

type fakeCommentNotifier struct {
	notices []notify.ReplyNotice
}

func (f *fakeCommentNotifier) NotifyReply(_ context.Context, notice notify.ReplyNotice) error {
	f.notices = append(f.notices, notice)
	return nil
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
				addResp: &model.CommentSchema{CommentID: tt.findResp.ID},
				findResp: map[int64]*model.Comment{
					tt.findResp.ID: tt.findResp,
				},
			}
			l := NewAddCommentLogic(context.Background(), &svc.ServiceContext{
				CommentModel:    stub,
				CommentNotifier: notify.NoopCommentNotifier{},
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

func TestAddCommentLogic_ValidateMessage(t *testing.T) {
	stub := &addCommentStubModel{}
	l := NewAddCommentLogic(context.Background(), &svc.ServiceContext{
		CommentModel:    stub,
		CommentNotifier: notify.NoopCommentNotifier{},
	})

	_, err := l.AddComment(&comment.CommentRequest{
		ObjID:    1001,
		ObjType:  1,
		MemberID: 2001,
		Message:  "   ",
	})
	if err == nil {
		t.Fatalf("expected error for empty message")
	}
	if got := int(status.Code(err)); got != 400 {
		t.Fatalf("status code=%d want=400", got)
	}

	_, err = l.AddComment(&comment.CommentRequest{
		ObjID:    1001,
		ObjType:  1,
		MemberID: 2001,
		Message:  strings.Repeat("a", types.MaxCommentLength+1),
	})
	if err == nil {
		t.Fatalf("expected error for too long message")
	}
	if got := int(status.Code(err)); got != 400 {
		t.Fatalf("status code=%d want=400", got)
	}
}

func TestAddCommentLogic_NotifyReply(t *testing.T) {
	now := time.Now()
	stub := &addCommentStubModel{
		addResp: &model.CommentSchema{CommentID: 9002},
		findResp: map[int64]*model.Comment{
			9002: {
				ID:        9002,
				ObjID:     1001,
				ObjType:   1,
				MemberID:  2002,
				RootID:    9001,
				ReplyID:   9001,
				Message:   "reply body",
				CreatedAt: now,
			},
			9001: {
				ID:        9001,
				ObjID:     1001,
				ObjType:   1,
				MemberID:  2001,
				Message:   "root body",
				CreatedAt: now,
			},
		},
	}
	notifier := &fakeCommentNotifier{}
	l := NewAddCommentLogic(context.Background(), &svc.ServiceContext{
		Config: config.Config{
			ReplyNoticeScope: struct {
				Domain      string
				TenantID    string
				ProjectID   string
				Environment string
			}{
				Domain:      "tenant",
				TenantID:    "t-1",
				ProjectID:   "p-1",
				Environment: "prod",
			},
		},
		CommentModel:    stub,
		CommentNotifier: notifier,
	})

	_, err := l.AddComment(&comment.CommentRequest{
		ObjID:    1001,
		ObjType:  1,
		MemberID: 2002,
		Message:  " reply body ",
		RootID:   9001,
		ReplyID:  9001,
	})
	if err != nil {
		t.Fatalf("AddComment() error = %v", err)
	}
	if len(notifier.notices) != 1 {
		t.Fatalf("len(notifier.notices) = %d, want 1", len(notifier.notices))
	}
	if notifier.notices[0].ReceiverID != 2001 {
		t.Fatalf("ReceiverID = %d, want 2001", notifier.notices[0].ReceiverID)
	}
	if notifier.notices[0].Domain != "tenant" {
		t.Fatalf("Domain = %q, want tenant", notifier.notices[0].Domain)
	}
	if notifier.notices[0].TenantID != "t-1" || notifier.notices[0].ProjectID != "p-1" || notifier.notices[0].Environment != "prod" {
		t.Fatalf("scope = %+v, want tenant/project/environment populated", notifier.notices[0])
	}
}
