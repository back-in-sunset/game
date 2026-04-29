package logic

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeMobileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangeMobileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeMobileLogic {
	return &ChangeMobileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeMobileLogic) ChangeMobile(req *types.ChangeMobileRequest) (*types.ChangeMobileResponse, error) {
	uid, err := l.ctx.Value(types.UserIDKey).(json.Number).Int64()
	if err != nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeUIDInvalid, "uid invalid")
	}
	req.NewMobile = strings.TrimSpace(req.NewMobile)
	req.VerifyCode = strings.TrimSpace(req.VerifyCode)
	if req.NewMobile == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeNewMobileRequired, "newMobile is required")
	}
	if req.VerifyCode == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeVerifyCodeRequired, "verifyCode is required")
	}
	if l.svcCtx.SmsVerifier == nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeSMSVerifierUnavailable, "sms verifier unavailable")
	}
	if err = l.svcCtx.SmsVerifier.VerifyCode(l.ctx, req.NewMobile, req.VerifyCode); err != nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeVerifyCodeInvalidOrExpired, "verification code invalid or expired")
	}

	if _, err = l.svcCtx.UserModel.FindOneByMobile(l.ctx, req.NewMobile); err == nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeMobileAlreadyInUse, "mobile already in use")
	} else if err != model.ErrNotFound {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeMobileIndexQueryFailed, err.Error())
	}

	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uid)
	if err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeUserQueryFailed, err.Error())
	}
	if u.Mobile == req.NewMobile {
		return &types.ChangeMobileResponse{Success: true, Message: "ok"}, nil
	}
	u.Mobile = req.NewMobile
	if err = l.svcCtx.UserModel.Update(l.ctx, u); err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeUserUpdateFailed, err.Error())
	}
	return &types.ChangeMobileResponse{Success: true, Message: "ok"}, nil
}
