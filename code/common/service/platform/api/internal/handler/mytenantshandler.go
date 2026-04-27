package handler

import (
	"net/http"

	"platform/api/internal/logic"
	"platform/api/internal/svc"
	"platform/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func MyTenantsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MyTenantsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewMyTenantsLogic(r.Context(), svcCtx)
		resp, err := l.MyTenants(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
