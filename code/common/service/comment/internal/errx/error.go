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
	return &BizError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
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
		if be.HTTPStatus <= 0 {
			be.HTTPStatus = http.StatusBadRequest
		}
		if be.Code == "" {
			be.Code = CodeBadRequestDefault
		}
		if be.Message == "" {
			be.Message = "bad request"
		}
		return be
	}

	if st, ok := status.FromError(err); ok {
		msg := st.Message()
		code := int(st.Code())
		if parsed := parseCodedMessage(code, msg); parsed != nil {
			return parsed
		}
		if mapped := mapKnownMessage(msg); mapped != nil {
			return mapped
		}
		switch {
		case code >= 400 && code < 500:
			return New(code, CodeBadRequestDefault, msg)
		case code >= 500 && code < 600:
			return New(code, CodeInternalDefault, msg)
		default:
			return New(http.StatusInternalServerError, CodeInternalDefault, msg)
		}
	}

	if mapped := mapKnownMessage(err.Error()); mapped != nil {
		return mapped
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

func mapKnownMessage(msg string) *BizError {
	switch msg {
	case "obj_id is required", "obj_id不能为空":
		return New(http.StatusBadRequest, CodeObjIDRequired, msg)
	case "obj_type is required", "obj_type不能为空":
		return New(http.StatusBadRequest, CodeObjTypeRequired, msg)
	case "member_id is required", "member_id不能为空":
		return New(http.StatusBadRequest, CodeMemberIDRequired, msg)
	case "comment_id is required", "comment_id不能为空":
		return New(http.StatusBadRequest, CodeCommentIDRequired, msg)
	case "message is required", "message不能为空":
		return New(http.StatusBadRequest, CodeMessageRequired, msg)
	case "message长度不能超过1000":
		return New(http.StatusBadRequest, CodeMessageTooLong, msg)
	case "sort type is invalid":
		return New(http.StatusBadRequest, CodeSortTypeInvalid, msg)
	case "无权限删除该评论", "permission denied":
		return New(http.StatusForbidden, CodePermissionDenied, msg)
	case "评论不存在":
		return New(http.StatusNotFound, CodeCommentNotFound, msg)
	case "invalid reply relation":
		return New(http.StatusBadRequest, CodeInvalidReply, msg)
	default:
		return nil
	}
}
