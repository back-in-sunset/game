package handler

import (
	"net/http"

	"history/api/internal/logic"
	"history/api/internal/svc"
	"history/api/internal/types"
	"history/internal/errx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RecordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RecordHistoryRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		l := logic.NewRecordLogic(r.Context(), svcCtx)
		resp, err := l.Record(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
