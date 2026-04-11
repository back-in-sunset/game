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

type UpdateShadowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateShadowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateShadowLogic {
	return &UpdateShadowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateShadowLogic) UpdateShadow(in *device_shadow.UpdateShadowRequest) (*device_shadow.UpdateShadowResponse, error) {
	shadow, err := l.svcCtx.ShadowModel.FindOne(l.ctx, in.DeviceId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(100, "device shadow not found")
		}
		return nil, status.Error(500, err.Error())
	}

	var desired, reported map[string]string
	if shadow.DesiredState.Valid {
		json.Unmarshal([]byte(shadow.DesiredState.String), &desired)
	}
	if shadow.ReportedState.Valid {
		json.Unmarshal([]byte(shadow.ReportedState.String), &reported)
	}

	for k, v := range in.Desired {
		desired[k] = v
	}

	delta := calculateDelta(desired, reported)

	desiredJSON, _ := json.Marshal(desired)
	deltaJSON, _ := json.Marshal(delta)

	shadow.DesiredState.String = string(desiredJSON)
	shadow.DesiredState.Valid = true
	shadow.DeltaState.String = string(deltaJSON)
	shadow.DeltaState.Valid = true
	shadow.Version++

	err = l.svcCtx.ShadowModel.Update(l.ctx, shadow)
	if err != nil {
		return nil, status.Error(500, err.Error())
	}

	return &device_shadow.UpdateShadowResponse{
		Success: true,
		State: &device_shadow.ShadowState{
			Desired:  desired,
			Reported: reported,
			Delta:    delta,
			Version:  shadow.Version,
		},
	}, nil
}

func calculateDelta(desired, reported map[string]string) map[string]string {
	delta := make(map[string]string)
	for k, v := range desired {
		if rv, ok := reported[k]; !ok || rv != v {
			delta[k] = v
		}
	}
	return delta
}
