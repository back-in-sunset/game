package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
)

type unBlockCommentStubModel struct {
	setStateResp *model.Comment
	setStateErr  error
	findResp     *model.Comment
	findErr      error
}

func (m *unBlockCommentStubModel) AddComment(context.Context, *model.CommentSubject, *model.CommentIndex, *model.CommentContent) (*model.CommentSchema, error) {
	return nil, nil
}

func (m *unBlockCommentStubModel) DeleteComment(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *unBlockCommentStubModel) CommentListByObjID(context.Context, int64, int64, int64, int64, string, int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *unBlockCommentStubModel) FindOneByObjID(context.Context, int64, int64) (*model.Comment, error) {
	return m.findResp, m.findErr
}

func (m *unBlockCommentStubModel) CacheCommentsByIDs(context.Context, int64, []int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *unBlockCommentStubModel) AdjustCommentLikeCount(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (m *unBlockCommentStubModel) SetCommentState(context.Context, int64, int64, int64) (*model.Comment, error) {
	return m.setStateResp, m.setStateErr
}

func (m *unBlockCommentStubModel) SetCommentAttrs(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func TestUnBlockCommentLogic_UnblocksComment(t *testing.T) {
	objID := time.Now().UnixNano()
	commentID := int64(9004)
	memberID := int64(1001)

	stub := &unBlockCommentStubModel{
		setStateResp: &model.Comment{
			ID:      commentID,
			ObjID:   objID,
			MemberID: memberID,
			State:   0,
		},
	}

	serviceCtx := &svc.ServiceContext{
		CommentModel: stub,
	}

	logic := NewUnBlockCommentLogic(context.Background(), serviceCtx)

	unblockResp, err := logic.UnBlockComment(&comment.UnBlockCommentRequest{
		ObjID:     objID,
		ObjType:   1,
		CommentID: commentID,
		MemberID:  memberID,
	})

	if err != nil {
		t.Fatalf("UnBlockComment() error = %v", err)
	}

	if !unblockResp.Success {
		t.Errorf("expected success=true, got %v", unblockResp.Success)
	}

	if unblockResp.Message != "ok" {
		t.Errorf("expected message='ok', got %v", unblockResp.Message)
	}
}

func TestUnBlockCommentLogic_NotFound(t *testing.T) {
	objID := time.Now().UnixNano()
	commentID := int64(9999)
	memberID := int64(1001)

	stub := &unBlockCommentStubModel{
		setStateErr: model.ErrNotFound,
	}

	serviceCtx := &svc.ServiceContext{
		CommentModel: stub,
	}

	logic := NewUnBlockCommentLogic(context.Background(), serviceCtx)

	_, err := logic.UnBlockComment(&comment.UnBlockCommentRequest{
		ObjID:     objID,
		ObjType:   1,
		CommentID: commentID,
		MemberID:  memberID,
	})

	if err == nil {
		t.Fatalf("expected error for non-existent comment, got nil")
	}
}

func TestUnBlockCommentLogic_Validation(t *testing.T) {
	serviceCtx := &svc.ServiceContext{
		CommentModel: &unBlockCommentStubModel{},
	}

	logic := NewUnBlockCommentLogic(context.Background(), serviceCtx)

	// Test missing ObjID
	_, err := logic.UnBlockComment(&comment.UnBlockCommentRequest{
		ObjType:   1,
		CommentID: 1,
		MemberID:  1,
	})
	if err == nil {
		t.Fatalf("expected error for missing ObjID, got nil")
	}

	// Test missing CommentID
	_, err = logic.UnBlockComment(&comment.UnBlockCommentRequest{
		ObjID:    1,
		ObjType:  1,
		MemberID: 1,
	})
	if err == nil {
		t.Fatalf("expected error for missing CommentID, got nil")
	}
}