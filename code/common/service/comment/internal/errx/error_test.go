package errx

import (
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNormalizeError_KnownPlainError(t *testing.T) {
	got := normalizeError(errors.New("obj_id is required"))
	if got.HTTPStatus != http.StatusBadRequest || got.Code != CodeObjIDRequired {
		t.Fatalf("unexpected error mapping: %+v", got)
	}
}

func TestNormalizeError_KnownGrpcError(t *testing.T) {
	got := normalizeError(status.Error(codes.Code(404), "评论不存在"))
	if got.HTTPStatus != http.StatusNotFound || got.Code != CodeCommentNotFound {
		t.Fatalf("unexpected grpc error mapping: %+v", got)
	}
}

func TestNormalizeError_CodedGrpcError(t *testing.T) {
	got := normalizeError(RPCError(http.StatusBadRequest, CodeInvalidReply, "invalid reply relation"))
	if got.HTTPStatus != http.StatusBadRequest || got.Code != CodeInvalidReply || got.Message != "invalid reply relation" {
		t.Fatalf("unexpected coded grpc error mapping: %+v", got)
	}
}

func TestNormalizeError_PermissionDenied(t *testing.T) {
	got := normalizeError(errors.New("permission denied"))
	if got.HTTPStatus != http.StatusForbidden || got.Code != CodePermissionDenied {
		t.Fatalf("unexpected permission mapping: %+v", got)
	}
}

func TestNormalizeError_UnknownError(t *testing.T) {
	got := normalizeError(errors.New("db down"))
	if got.HTTPStatus != http.StatusInternalServerError || got.Code != CodeInternalDefault {
		t.Fatalf("unexpected unknown error mapping: %+v", got)
	}
}
