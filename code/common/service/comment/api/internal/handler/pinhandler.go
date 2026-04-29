package handler

import (
	"net/http"

	"comment/api/internal/logic"
	"comment/api/internal/svc"
	"comment/api/internal/types"
	"comment/internal/errx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PinHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CommentActionRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}

		l := logic.NewPinLogic(r.Context(), svcCtx)
		resp, err := l.Pin(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
