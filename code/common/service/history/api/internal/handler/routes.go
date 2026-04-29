package handler

import (
	"net/http"

	"history/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/history/record", Handler: RecordHandler(serverCtx)},
		{Method: http.MethodGet, Path: "/api/history/list", Handler: ListHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/history/item", Handler: DeleteItemHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/history/type", Handler: ClearByTypeHandler(serverCtx)},
		{Method: http.MethodDelete, Path: "/api/history/all", Handler: ClearAllHandler(serverCtx)},
	})
}
