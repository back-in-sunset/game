package logic

import (
	"context"
	"encoding/json"

	"game/server/device_shadow/rpc/device_shadow"
	"game/server/device_shadow/rpc/internal/svc"
	"game/server/device_shadow/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type GetDeltaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeltaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeltaLogic {
	return &GetDeltaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeltaLogic) GetDelta(in *device_shadow.GetDeltaRequest) (*device_shadow.GetDeltaResponse, error) {
	shadow, err := l.svcCtx.ShadowModel.FindOne(l.ctx, in.DeviceId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(100, "device shadow not found")
		}
		return nil, status.Error(500, err.Error())
	}

	var delta map[string]string
	if shadow.DeltaState.Valid {
		json.Unmarshal([]byte(shadow.DeltaState.String), &delta)
	}

	return &device_shadow.GetDeltaResponse{
		DeviceId:        shadow.DeviceId,
		Delta:           delta,
		DesiredVersion:  shadow.Version,
		ReportedVersion: shadow.Version,
	}, nil
}
