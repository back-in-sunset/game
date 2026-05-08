package errx

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var codeToGRPC = map[string]codes.Code{
	F1001: codes.NotFound,
	F1002: codes.InvalidArgument,
	F1003: codes.AlreadyExists,
	F1004: codes.NotFound,
	F1005: codes.FailedPrecondition,
	F1006: codes.NotFound,
	F1007: codes.PermissionDenied,
	F1008: codes.AlreadyExists,
	F1009: codes.AlreadyExists,
	F1010: codes.NotFound,
}

func ToGRPC(code string) error {
	c, ok := codeToGRPC[code]
	if !ok {
		c = codes.Internal
	}
	return status.Error(c, code)
}

func Is(err error, code string) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	return st.Message() == code
}
