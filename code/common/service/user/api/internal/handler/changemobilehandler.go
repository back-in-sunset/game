package handler

import (
	"net/http"

	"user/api/internal/errx"
	"user/api/internal/logic"
	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ChangeMobileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChangeMobileRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}

		l := logic.NewChangeMobileLogic(r.Context(), svcCtx)
		resp, err := l.ChangeMobile(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
