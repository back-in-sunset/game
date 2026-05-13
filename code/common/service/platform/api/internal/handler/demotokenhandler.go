package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"platform/api/internal/logic"
	"platform/api/internal/svc"
	"platform/api/internal/types"
)

func DemoTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DemoTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := logic.NewDemoTokenLogic(r.Context(), svcCtx)
		resp, err := l.DemoToken(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
