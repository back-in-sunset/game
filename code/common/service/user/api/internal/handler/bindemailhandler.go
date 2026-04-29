package handler

import (
	"net/http"

	"user/api/internal/errx"
	"user/api/internal/logic"
	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func BindEmailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindEmailRequest
		if err := httpx.Parse(r, &req); err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
			return
		}

		l := logic.NewBindEmailLogic(r.Context(), svcCtx)
		resp, err := l.BindEmail(&req)
		if err != nil {
			errx.WriteHTTPError(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
