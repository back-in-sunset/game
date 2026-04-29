package errx

import (
	"net/http"
	"testing"
)

func TestNormalizeError_CodedGrpcError(t *testing.T) {
	got := normalizeError(RPCError(http.StatusBadRequest, CodeMediaTypeInvalid, "media_type invalid"))
	if got.HTTPStatus != http.StatusBadRequest || got.Code != CodeMediaTypeInvalid || got.Message != "media_type invalid" {
		t.Fatalf("unexpected error mapping: %+v", got)
	}
}
