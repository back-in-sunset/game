package errx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BizError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *BizError) Error() string {
	return e.Message
}

func New(httpStatus int, code string, message string) *BizError {
	return &BizError{HTTPStatus: httpStatus, Code: code, Message: message}
}

func RPCError(httpStatus int, code string, message string) error {
	return status.Error(codes.Code(httpStatus), code+"|"+message)
}

func WriteHTTPError(_ context.Context, w http.ResponseWriter, err error) {
	biz := normalizeError(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(biz.HTTPStatus)
	_ = json.NewEncoder(w).Encode(biz)
}

func normalizeError(err error) *BizError {
	if err == nil {
		return New(http.StatusInternalServerError, CodeInternalDefault, "internal error")
	}
	var be *BizError
	if errors.As(err, &be) {
		return be
	}
	if st, ok := status.FromError(err); ok {
		if parsed := parseCodedMessage(int(st.Code()), st.Message()); parsed != nil {
			return parsed
		}
		return New(http.StatusInternalServerError, CodeInternalDefault, st.Message())
	}
	return New(http.StatusInternalServerError, CodeInternalDefault, err.Error())
}

func parseCodedMessage(httpStatus int, msg string) *BizError {
	parts := strings.SplitN(msg, "|", 2)
	if len(parts) != 2 || parts[0] == "" {
		return nil
	}
	if httpStatus < 100 || httpStatus > 599 {
		httpStatus = http.StatusBadRequest
	}
	return New(httpStatus, parts[0], parts[1])
}
