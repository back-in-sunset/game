package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"

	"google.golang.org/grpc/status"
)

type deleteCommentStubModel struct {
	deleteResp *model.Comment
	deleteErr  error
}

func (m *deleteCommentStubModel) AddComment(context.Context, *model.CommentSubject, *model.CommentIndex, *model.CommentContent) (*model.CommentSchema, error) {
	return nil, nil
}

func (m *deleteCommentStubModel) DeleteComment(context.Context, int64, int64, int64) (*model.Comment, error) {
	return m.deleteResp, m.deleteErr
}

func (m *deleteCommentStubModel) CommentListByObjID(context.Context, int64, int64, int64, int64, string, int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *deleteCommentStubModel) FindOneByObjID(context.Context, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *deleteCommentStubModel) CacheCommentsByIDs(context.Context, int64, []int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *deleteCommentStubModel) AdjustCommentLikeCount(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (m *deleteCommentStubModel) SetCommentState(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *deleteCommentStubModel) SetCommentAttrs(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func TestDeleteCommentLogic_PermissionDenied(t *testing.T) {
	stub := &deleteCommentStubModel{
		deleteErr: model.ErrPermissionDenied,
	}
	l := NewDeleteCommentLogic(context.Background(), &svc.ServiceContext{
		CommentModel: stub,
	})

	_, err := l.DeleteComment(&comment.CommentRequest{
		ObjID:     1001,
		ObjType:   1,
		CommentID: 9001,
		MemberID:  2002,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := int(status.Code(err)); got != 403 {
		t.Fatalf("status code=%d want=403", got)
	}
}

func TestDeleteCommentLogic_NotFoundAndInvalid(t *testing.T) {
	l := NewDeleteCommentLogic(context.Background(), &svc.ServiceContext{
		CommentModel: &deleteCommentStubModel{deleteErr: model.ErrNotFound},
	})

	_, err := l.DeleteComment(&comment.CommentRequest{
		ObjID:     1001,
		ObjType:   1,
		CommentID: 9001,
		MemberID:  2001,
	})
	if err == nil {
		t.Fatalf("expected not found error")
	}
	if got := int(status.Code(err)); got != 404 {
		t.Fatalf("status code=%d want=404", got)
	}

	_, err = l.DeleteComment(&comment.CommentRequest{
		ObjID:     0,
		CommentID: 9001,
		MemberID:  2001,
	})
	if err == nil {
		t.Fatalf("expected invalid obj_id error")
	}
	if got := int(status.Code(err)); got != 400 {
		t.Fatalf("status code=%d want=400", got)
	}
}

func TestDeleteCommentLogic_Success(t *testing.T) {
	stub := &deleteCommentStubModel{
		deleteResp: &model.Comment{
			ID:        9001,
			ObjID:     1001,
			ObjType:   1,
			MemberID:  2001,
			Message:   "ok",
			CreatedAt: time.Now(),
		},
	}
	l := NewDeleteCommentLogic(context.Background(), &svc.ServiceContext{
		CommentModel: stub,
	})

	resp, err := l.DeleteComment(&comment.CommentRequest{
		ObjID:     1001,
		ObjType:   1,
		CommentID: 9001,
		MemberID:  2001,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CommentID != 9001 {
		t.Fatalf("resp.CommentID=%d want=9001", resp.CommentID)
	}
}

func TestDeleteCommentLogic_InternalError(t *testing.T) {
	l := NewDeleteCommentLogic(context.Background(), &svc.ServiceContext{
		CommentModel: &deleteCommentStubModel{deleteErr: errors.New("db down")},
	})

	_, err := l.DeleteComment(&comment.CommentRequest{
		ObjID:     1001,
		ObjType:   1,
		CommentID: 9001,
		MemberID:  2001,
	})
	if err == nil {
		t.Fatalf("expected internal error")
	}
	if got := int(status.Code(err)); got != 500 {
		t.Fatalf("status code=%d want=500", got)
	}
}
