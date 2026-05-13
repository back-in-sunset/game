package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
)

type blockCommentStubModel struct {
	setStateResp *model.Comment
	setStateErr  error
	findResp     *model.Comment
	findErr      error
}

func (m *blockCommentStubModel) AddComment(context.Context, *model.CommentSubject, *model.CommentIndex, *model.CommentContent) (*model.CommentSchema, error) {
	return nil, nil
}

func (m *blockCommentStubModel) DeleteComment(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func (m *blockCommentStubModel) CommentListByObjID(context.Context, int64, int64, int64, int64, string, int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *blockCommentStubModel) FindOneByObjID(context.Context, int64, int64) (*model.Comment, error) {
	return m.findResp, m.findErr
}

func (m *blockCommentStubModel) CacheCommentsByIDs(context.Context, int64, []int64) ([]*model.Comment, error) {
	return nil, nil
}

func (m *blockCommentStubModel) AdjustCommentLikeCount(context.Context, int64, int64, int64) (int64, error) {
	return 0, nil
}

func (m *blockCommentStubModel) SetCommentState(context.Context, int64, int64, int64) (*model.Comment, error) {
	return m.setStateResp, m.setStateErr
}

func (m *blockCommentStubModel) SetCommentAttrs(context.Context, int64, int64, int64) (*model.Comment, error) {
	return nil, model.ErrNotFound
}

func TestBlockCommentLogic_BlocksComment(t *testing.T) {
	objID := time.Now().UnixNano()
	commentID := int64(9003)
	memberID := int64(1001)

	stub := &blockCommentStubModel{
		setStateResp: &model.Comment{
			ID:      commentID,
			ObjID:   objID,
			MemberID: memberID,
			State:   1,
		},
	}

	serviceCtx := &svc.ServiceContext{
		CommentModel: stub,
	}

	logic := NewBlockCommentLogic(context.Background(), serviceCtx)

	blockResp, err := logic.BlockComment(&comment.BlockCommentRequest{
		ObjID:     objID,
		ObjType:   1,
		CommentID: commentID,
		MemberID:  memberID,
	})

	if err != nil {
		t.Fatalf("BlockComment() error = %v", err)
	}

	if !blockResp.Success {
		t.Errorf("expected success=true, got %v", blockResp.Success)
	}

	if blockResp.Message != "ok" {
		t.Errorf("expected message='ok', got %v", blockResp.Message)
	}
}

func TestBlockCommentLogic_NotFound(t *testing.T) {
	objID := time.Now().UnixNano()
	commentID := int64(9999)
	memberID := int64(1001)

	stub := &blockCommentStubModel{
		setStateErr: model.ErrNotFound,
	}

	serviceCtx := &svc.ServiceContext{
		CommentModel: stub,
	}

	logic := NewBlockCommentLogic(context.Background(), serviceCtx)

	_, err := logic.BlockComment(&comment.BlockCommentRequest{
		ObjID:     objID,
		ObjType:   1,
		CommentID: commentID,
		MemberID:  memberID,
	})

	if err == nil {
		t.Fatalf("expected error for non-existent comment, got nil")
	}
}

func TestBlockCommentLogic_Validation(t *testing.T) {
	serviceCtx := &svc.ServiceContext{
		CommentModel: &blockCommentStubModel{},
	}

	logic := NewBlockCommentLogic(context.Background(), serviceCtx)

	// Test missing ObjID
	_, err := logic.BlockComment(&comment.BlockCommentRequest{
		ObjType:   1,
		CommentID: 1,
		MemberID:  1,
	})
	if err == nil {
		t.Fatalf("expected error for missing ObjID, got nil")
	}

	// Test missing CommentID
	_, err = logic.BlockComment(&comment.BlockCommentRequest{
		ObjID:    1,
		ObjType:  1,
		MemberID: 1,
	})
	if err == nil {
		t.Fatalf("expected error for missing CommentID, got nil")
	}
}