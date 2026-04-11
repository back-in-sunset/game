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

type GetShadowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShadowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShadowLogic {
	return &GetShadowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetShadowLogic) GetShadow(in *device_shadow.GetShadowRequest) (*device_shadow.GetShadowResponse, error) {
	shadow, err := l.svcCtx.ShadowModel.FindOne(l.ctx, in.DeviceId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(100, "device shadow not found")
		}
		return nil, status.Error(500, err.Error())
	}

	var desired, reported, delta map[string]string
	if shadow.DesiredState.Valid {
		json.Unmarshal([]byte(shadow.DesiredState.String), &desired)
	}
	if shadow.ReportedState.Valid {
		json.Unmarshal([]byte(shadow.ReportedState.String), &reported)
	}
	if shadow.DeltaState.Valid {
		json.Unmarshal([]byte(shadow.DeltaState.String), &delta)
	}

	return &device_shadow.GetShadowResponse{
		DeviceId:   shadow.DeviceId,
		DeviceName: shadow.DeviceName,
		State: &device_shadow.ShadowState{
			Desired:  desired,
			Reported: reported,
			Delta:    delta,
			Version:  shadow.Version,
		},
		CreatedAt: shadow.CreatedAt,
		UpdatedAt: shadow.UpdatedAt,
	}, nil
}
