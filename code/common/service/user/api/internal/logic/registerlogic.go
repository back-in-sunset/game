package logic

import (
	"context"
	"net/http"
	"strings"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/api/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Mobile = strings.TrimSpace(req.Mobile)
	if req.Name == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeNameRequired, "name is required")
	}
	if req.Mobile == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeMobileRequired, "mobile is required")
	}
	if req.Password == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodePasswordRequired, "password is required")
	}

	res, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterRequest{
		Name:     req.Name,
		Gender:   req.Gender,
		Mobile:   req.Mobile,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &types.RegisterResponse{
		ID:     res.ID,
		Name:   res.Name,
		Gender: res.Gender,
		Mobile: res.Mobile,
	}, nil
}
