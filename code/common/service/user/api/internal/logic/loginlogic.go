package logic

import (
	"context"
	"net/http"
	"strings"
	"time"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/api/userclient"
	"user/utils/jwtx"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	req.Mobile = strings.TrimSpace(req.Mobile)
	if req.Mobile == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeMobileRequired, "mobile is required")
	}
	if req.Password == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodePasswordRequired, "password is required")
	}

	res, err := l.svcCtx.UserRpc.Login(l.ctx, &userclient.LoginRequest{
		Mobile:   req.Mobile,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire

	accessToken, err := jwtx.GetToken(l.svcCtx.PrivateKey, res.ID, accessExpire)
	if err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeLoginFailed, err.Error())
	}

	return &types.LoginResponse{
		AccessToken:  accessToken,
		AccessExpire: now + accessExpire,
	}, nil
}
