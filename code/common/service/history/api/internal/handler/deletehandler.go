package handler

import (
	"net/http"

	"history/api/internal/logic"
	"history/api/internal/svc"
	"history/api/internal/types"
	"history/internal/errx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteItemHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteHistoryItemRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		l := logic.NewDeleteLogic(r.Context(), svcCtx)
		resp, err := l.DeleteItem(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func ClearByTypeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClearHistoryByTypeRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		l := logic.NewDeleteLogic(r.Context(), svcCtx)
		resp, err := l.ClearByType(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func ClearAllHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ClearHistoryAllRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		l := logic.NewDeleteLogic(r.Context(), svcCtx)
		resp, err := l.ClearAll(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
