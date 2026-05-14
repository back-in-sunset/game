# Comment Advanced Capabilities Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add comprehensive tests for comment advanced capabilities (like/unlike/block/unblock/pin/unpin) and verify filtering/sorting works correctly

**Architecture:** Most features are already implemented. This plan focuses on adding integration tests and verifying the filtering logic for blocked/pinned comments.

**Tech Stack:** Go, go-zero, MySQL, Redis, testcontainers

---

## Context

**Current State:**
- Backend RPC: Like/Unlike/Block/Unblock/SetCommentAttrs (for pin) all implemented
- Backend API: All handlers and routes exist in `code/common/service/comment/api/internal/handler/routes.go`
- Frontend: CommentPage.tsx (661 lines) has complete UI and handlers
- Frontend API: All methods exported through `@game/api` package

**What's Missing:**
1. No integration tests for block/unblock/pin/unpin operations
2. Need to verify blocked comments are filtered from normal list
3. Need to verify pinned comments appear at top of sorted list

**Key Files:**
- RPC Logic: `code/common/service/comment/rpc/internal/logic/`
- API Logic: `code/common/service/comment/api/internal/logic/`
- API Routes: `code/common/service/comment/api/internal/handler/routes.go`
- Existing Test Pattern: `code/common/service/comment/rpc/internal/logic/likecommentlogic_test.go`

---

### Task 1: Add block/unblock integration tests

**Files:**
- Create: `code/common/service/comment/rpc/internal/logic/blockcommentlogic_test.go`
- Create: `code/common/service/comment/rpc/internal/logic/unblockcommentlogic_test.go`
- Test Pattern Reference: `code/common/service/comment/rpc/internal/logic/likecommentlogic_test.go`

- [ ] **Step 1: Write the failing test for block**

```go
package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
	"comment/rpc/model"
)

func TestBlockCommentLogic_BlocksComment(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable")
	}

	rds := mustTestRedis(t)
	serviceCtx := &svc.ServiceContext{Redis: rds}
	logic := NewBlockCommentLogic(context.Background(), serviceCtx)

	objID := time.Now().UnixNano()
	objType := int64(1)
	memberID := int64(1001)

	// Create a test comment first
	createLogic := NewAddCommentLogic(context.Background(), serviceCtx)
	commentResp, err := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Test comment to block",
	})
	if err != nil {
		t.Fatalf("create test comment: %v", err)
	}

	// Block the comment
	blockResp, err := logic.BlockComment(&comment.BlockCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentResp.CommentID,
	})
	if err != nil {
		t.Fatalf("block comment: %v", err)
	}

	if !blockResp.Success {
		t.Errorf("expected success=true, got %v", blockResp.Success)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run TestBlockCommentLogic_BlocksComment`
Expected: FAIL with "function not defined" or other errors

- [ ] **Step 3: Check if BlockCommentLogic already exists**

Run: `ls code/common/service/comment/rpc/internal/logic/blockcommentlogic.go`
Expected: File exists (feature already implemented)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run TestBlockCommentLogic_BlocksComment`
Expected: PASS if logic exists, implement if missing

- [ ] **Step 5: Add test for unblock**

```go
func TestUnBlockCommentLogic_UnblocksComment(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable")
	}

	rds := mustTestRedis(t)
	serviceCtx := &svc.ServiceContext{Redis: rds}
	logic := NewUnBlockCommentLogic(context.Background(), serviceCtx)

	objID := time.Now().UnixNano()
	objType := int64(1)
	memberID := int64(1001)

	// Create and block a comment first
	createLogic := NewAddCommentLogic(context.Background(), serviceCtx)
	commentResp, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Test comment to unblock",
	})
	NewBlockCommentLogic(context.Background(), serviceCtx).BlockComment(&comment.BlockCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentResp.CommentID,
	})

	// Unblock the comment
	unblockResp, err := logic.UnBlockComment(&comment.UnBlockCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentResp.CommentID,
	})
	if err != nil {
		t.Fatalf("unblock comment: %v", err)
	}

	if !unblockResp.Success {
		t.Errorf("expected success=true, got %v", unblockResp.Success)
	}
}
```

- [ ] **Step 6: Run tests and verify all pass**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run "TestBlockCommentLogic|TestUnBlockCommentLogic"`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add code/common/service/comment/rpc/internal/logic/blockcommentlogic_test.go code/common/service/comment/rpc/internal/logic/unblockcommentlogic_test.go
git commit -m "test: add block/unblock integration tests"
```

---

### Task 2: Add pin/unpin integration tests

**Files:**
- Create: `code/common/service/comment/rpc/internal/logic/setcommentattrslogic_test.go`
- Create: `code/common/service/comment/rpc/internal/logic/unsetcommentattrslogic_test.go`

- [ ] **Step 1: Write the failing test for pin**

```go
package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
)

func TestSetCommentAttrsLogic_PinsComment(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable")
	}

	rds := mustTestRedis(t)
	serviceCtx := &svc.ServiceContext{Redis: rds}
	logic := NewSetCommentAttrsLogic(context.Background(), serviceCtx)

	objID := time.Now().UnixNano()
	objType := int64(1)
	memberID := int64(1001)

	// Create a test comment first
	createLogic := NewAddCommentLogic(context.Background(), serviceCtx)
	commentResp, err := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Test comment to pin",
	})
	if err != nil {
		t.Fatalf("create test comment: %v", err)
	}

	// Pin the comment
	pinResp, err := logic.SetCommentAttrs(&comment.SetCommentAttrsRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentResp.CommentID,
		MemberID:  memberID,
	})
	if err != nil {
		t.Fatalf("pin comment: %v", err)
	}

	if !pinResp.Success {
		t.Errorf("expected success=true, got %v", pinResp.Success)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run TestSetCommentAttrsLogic_PinsComment`
Expected: FAIL with "function not defined" or pass if logic exists

- [ ] **Step 3: Write test for unpin**

```go
func TestUnSetCommentAttrsLogic_UnpinsComment(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable")
	}

	rds := mustTestRedis(t)
	serviceCtx := &svc.ServiceContext{Redis: rds}
	logic := NewUnSetCommentAttrsLogic(context.Background(), serviceCtx)

	objID := time.Now().UnixNano()
	objType := int64(1)
	memberID := int64(1001)

	// Create and pin a comment first
	createLogic := NewAddCommentLogic(context.Background(), serviceCtx)
	commentResp, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Test comment to unpin",
	})
	NewSetCommentAttrsLogic(context.Background(), serviceCtx).SetCommentAttrs(&comment.SetCommentAttrsRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentResp.CommentID,
		MemberID:  memberID,
	})

	// Unpin the comment
	unpinResp, err := logic.UnSetCommentAttrs(&comment.UnSetCommentAttrsRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: commentResp.CommentID,
		MemberID:  memberID,
	})
	if err != nil {
		t.Fatalf("unpin comment: %v", err)
	}

	if !unpinResp.Success {
		t.Errorf("expected success=true, got %v", unpinResp.Success)
	}
}
```

- [ ] **Step 4: Run tests and verify all pass**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run "TestSetCommentAttrsLogic|TestUnSetCommentAttrsLogic"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add code/common/service/comment/rpc/internal/logic/setcommentattrslogic_test.go code/common/service/comment/rpc/internal/logic/unsetcommentattrslogic_test.go
git commit -m "test: add pin/unpin integration tests"
```

---

### Task 3: Verify blocked comments are filtered from list

**Files:**
- Create: `code/common/service/comment/rpc/internal/logic/getcommentlistlogic_blocked_test.go`
- Reference: `code/common/service/comment/rpc/internal/logic/getcommentlistlogic_test.go`

- [ ] **Step 1: Write test for blocked comment filtering**

```go
package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
)

func TestGetCommentListLogic_BlockedCommentsHidden(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable")
	}

	rds := mustTestRedis(t)
	serviceCtx := &svc.ServiceContext{Redis: rds}
	logic := NewGetCommentListLogic(context.Background(), serviceCtx)

	objID := time.Now().UnixNano()
	objType := int64(1)
	memberID := int64(1001)

	// Create 3 comments
	createLogic := NewAddCommentLogic(context.Background(), serviceCtx)
	comment1, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Normal comment 1",
	})
	comment2, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Blocked comment",
	})
	comment3, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Normal comment 2",
	})

	// Block comment2
	NewBlockCommentLogic(context.Background(), serviceCtx).BlockComment(&comment.BlockCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: comment2.CommentID,
	})

	// Get list - blocked comment should be hidden
	listResp, err := logic.GetCommentList(&comment.GetCommentListRequest{
		ObjID:   objID,
		ObjType: objType,
	})
	if err != nil {
		t.Fatalf("get comment list: %v", err)
	}

	// Should only have 2 comments (comment2 is blocked)
	if len(listResp.List) != 2 {
		t.Errorf("expected 2 comments (1 blocked), got %d", len(listResp.List))
	}

	// Verify blocked comment is not in list
	commentIDs := make(map[int64]bool)
	for _, c := range listResp.List {
		commentIDs[c.CommentID] = true
	}
	if commentIDs[comment2.CommentID] {
		t.Error("blocked comment should not be in list")
	}
	if !commentIDs[comment1.CommentID] || !commentIDs[comment3.CommentID] {
		t.Error("normal comments should be in list")
	}
}
```

- [ ] **Step 2: Run test to verify behavior**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run TestGetCommentListLogic_BlockedCommentsHidden`
Expected: PASS if filtering works, FAIL if blocked comments appear

- [ ] **Step 3: If test fails, check getcommentlistlogic.go for filtering logic**

Run: `grep -n "state\|State" code/common/service/comment/rpc/internal/logic/getcommentlistlogic.go`
Expected: Should see filtering by state field

- [ ] **Step 4: Commit**

```bash
git add code/common/service/comment/rpc/internal/logic/getcommentlistlogic_blocked_test.go
git commit -m "test: verify blocked comments are filtered from list"
```

---

### Task 4: Verify pinned comments sort correctly

**Files:**
- Create: `code/common/service/comment/rpc/internal/logic/getcommentlistlogic_pinned_test.go`

- [ ] **Step 1: Write test for pinned comment sorting**

```go
package logic

import (
	"context"
	"testing"
	"time"

	"comment/rpc/comment"
	"comment/rpc/internal/svc"
)

func TestGetCommentListLogic_PinnedCommentsFirst(t *testing.T) {
	if !portReachable("127.0.0.1:6379") {
		t.Skip("skip integration test: redis is unavailable")
	}

	rds := mustTestRedis(t)
	serviceCtx := &svc.ServiceContext{Redis: rds}
	logic := NewGetCommentListLogic(context.Background(), serviceCtx)

	objID := time.Now().UnixNano()
	objType := int64(1)
	memberID := int64(1001)

	// Create 3 comments
	createLogic := NewAddCommentLogic(context.Background(), serviceCtx)
	comment1, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "First comment",
	})
	time.Sleep(10 * time.Millisecond)
	comment2, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Pinned comment",
	})
	time.Sleep(10 * time.Millisecond)
	comment3, _ := createLogic.AddComment(&comment.AddCommentRequest{
		ObjID:     objID,
		ObjType:   objType,
		MemberID:  memberID,
		Message:   "Last comment",
	})

	// Pin comment2
	NewSetCommentAttrsLogic(context.Background(), serviceCtx).SetCommentAttrs(&comment.SetCommentAttrsRequest{
		ObjID:     objID,
		ObjType:   objType,
		CommentID: comment2.CommentID,
		MemberID:  memberID,
	})

	// Get list sorted by time - pinned comment should be first
	listResp, err := logic.GetCommentList(&comment.GetCommentListRequest{
		ObjID:   objID,
		ObjType: objType,
	})
	if err != nil {
		t.Fatalf("get comment list: %v", err)
	}

	if len(listResp.List) < 3 {
		t.Fatalf("expected at least 3 comments, got %d", len(listResp.List))
	}

	// First comment should be pinned one
	if listResp.List[0].CommentID != comment2.CommentID {
		t.Errorf("expected pinned comment first, got comment id %d", listResp.List[0].CommentID)
	}
}
```

- [ ] **Step 2: Run test to verify behavior**

Run: `cd code/common/service/comment/rpc && go test -v ./internal/logic -run TestGetCommentListLogic_PinnedCommentsFirst`
Expected: PASS if pinned comments sort first

- [ ] **Step 3: Commit**

```bash
git add code/common/service/comment/rpc/internal/logic/getcommentlistlogic_pinned_test.go
git commit -m "test: verify pinned comments sort to top"
```

---

### Task 5: Run all comment tests and verify no regressions

**Files:**
- Test: All comment test files

- [ ] **Step 1: Run all comment RPC tests**

Run: `cd code/common/service/comment/rpc && go test -v ./...`
Expected: All tests pass

- [ ] **Step 2: Run all comment API tests**

Run: `cd code/common/service/comment/api && go test -v ./...`
Expected: All tests pass

- [ ] **Step 3: Create summary report**

Run: `cd code/common/service/comment/rpc && go test ./... -cover > coverage.txt`
Expected: Coverage report generated

- [ ] **Step 4: Commit any fixes if tests revealed issues**

```bash
git add code/common/service/comment/
git commit -m "test: fix issues revealed by comment tests"
```

---

## Completion Criteria

- [ ] All integration tests for like/unlike/block/unblock/pin/unpin pass
- [ ] Blocked comments are correctly filtered from list
- [ ] Pinned comments appear at top of sorted list
- [ ] No regression in existing comment functionality
- [ ] Test coverage for comment RPC is acceptable (>60%)