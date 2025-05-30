package logic

import (
	"context"

	"user/rpc/internal/svc"
	"user/rpc/pb/.user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *__user.LoginRequest) (*__user.LoginResponse, error) {
	// todo: add your logic here and delete this line

	return &__user.LoginResponse{}, nil
}
