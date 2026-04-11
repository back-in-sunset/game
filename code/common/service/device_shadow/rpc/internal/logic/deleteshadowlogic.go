package logic

import (
	"context"

	"game/server/device_shadow/rpc/device_shadow"
	"game/server/device_shadow/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type DeleteShadowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteShadowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteShadowLogic {
	return &DeleteShadowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteShadowLogic) DeleteShadow(in *device_shadow.DeleteShadowRequest) (*device_shadow.DeleteShadowResponse, error) {
	err := l.svcCtx.ShadowModel.Delete(l.ctx, in.DeviceId)
	if err != nil {
		return nil, status.Error(500, err.Error())
	}

	return &device_shadow.DeleteShadowResponse{
		Success: true,
	}, nil
}
