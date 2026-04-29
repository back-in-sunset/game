package errx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

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
		code := int(st.Code())
		msg := st.Message()
		switch {
		case code >= 400 && code < 500:
			return New(code, CodeBadRequestDefault, msg)
		case code >= 500 && code < 600:
			return New(code, CodeInternalDefault, msg)
		case code == 100:
			return New(http.StatusBadRequest, CodeRiskRejected, msg)
		case code == 403:
			return New(http.StatusForbidden, CodeUnauthorized, msg)
		default:
			return New(http.StatusBadRequest, CodeBadRequestDefault, msg)
		}
	}

	return New(http.StatusInternalServerError, CodeInternalDefault, err.Error())
}
