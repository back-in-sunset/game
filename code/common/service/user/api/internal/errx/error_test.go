package errx

import (
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNormalizeError_BizError(t *testing.T) {
	got := normalizeError(New(http.StatusBadRequest, CodeMobileRequired, "mobile is required"))
	if got.HTTPStatus != http.StatusBadRequest || got.Code != CodeMobileRequired || got.Message != "mobile is required" {
		t.Fatalf("unexpected biz error normalize result: %+v", got)
	}
}

func TestNormalizeError_GrpcPermissionDenied(t *testing.T) {
	got := normalizeError(status.Error(codes.PermissionDenied, "forbidden"))
	if got.HTTPStatus != http.StatusBadRequest || got.Code != CodeBadRequestDefault || got.Message != "forbidden" {
		t.Fatalf("unexpected grpc status normalize result: %+v", got)
	}
}

func TestNormalizeError_PlainError(t *testing.T) {
	got := normalizeError(errors.New("boom"))
	if got.HTTPStatus != http.StatusInternalServerError || got.Code != CodeInternalDefault || got.Message != "boom" {
		t.Fatalf("unexpected plain error normalize result: %+v", got)
	}
}
