package logic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/model"

	mysqlerr "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

type BindEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindEmailLogic {
	return &BindEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindEmailLogic) BindEmail(req *types.BindEmailRequest) (*types.BindEmailResponse, error) {
	uid, err := l.ctx.Value(types.UserIDKey).(json.Number).Int64()
	if err != nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeUIDInvalid, "uid invalid")
	}
	req.Email = strings.TrimSpace(req.Email)
	req.VerifyCode = strings.TrimSpace(req.VerifyCode)
	if req.Email == "" || req.VerifyCode == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeBindEmailRequired, "email and verifyCode are required")
	}
	if l.svcCtx.EmailVerifier == nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeEmailVerifierUnavailable, "email verifier unavailable")
	}
	if err = l.svcCtx.EmailVerifier.VerifyCode(l.ctx, req.Email, req.VerifyCode); err != nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeVerifyCodeInvalidOrExpired, "verification code invalid or expired")
	}

	exists, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err == nil && exists.UserID != uid {
		return nil, errx.New(http.StatusBadRequest, errx.CodeEmailAlreadyInUse, "email already in use")
	}
	if err != nil && err != model.ErrNotFound {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeEmailIndexQueryFailed, err.Error())
	}

	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uid)
	if err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeUserQueryFailed, err.Error())
	}
	u.Email = req.Email
	if err = l.svcCtx.UserModel.Update(l.ctx, u); err != nil {
		var me *mysqlerr.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return nil, errx.New(http.StatusBadRequest, errx.CodeEmailAlreadyInUse, "email already in use")
		}
		return nil, errx.New(http.StatusInternalServerError, errx.CodeUserUpdateFailed, err.Error())
	}
	return &types.BindEmailResponse{
		Success: true,
		Message: "ok",
	}, nil
}
