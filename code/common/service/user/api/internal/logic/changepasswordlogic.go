package logic

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/utils/cryptx"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (*types.ChangePasswordResponse, error) {
	uid, err := l.ctx.Value(types.UserIDKey).(json.Number).Int64()
	if err != nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeUIDInvalid, "uid invalid")
	}
	req.OldPassword = strings.TrimSpace(req.OldPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.OldPassword == "" || req.NewPassword == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeChangePasswordRequired, "oldPassword and newPassword are required")
	}
	if req.OldPassword == req.NewPassword {
		return nil, errx.New(http.StatusBadRequest, errx.CodePasswordMustDiffer, "old and new password must differ")
	}
	if len(req.NewPassword) < 6 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeNewPasswordTooShort, "new password length must be >= 6")
	}

	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uid)
	if err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeUserQueryFailed, err.Error())
	}
	oldEnc := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.OldPassword)
	if oldEnc != u.Password {
		return nil, errx.New(http.StatusBadRequest, errx.CodeOldPasswordIncorrect, "old password is incorrect")
	}

	u.Password = cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.NewPassword)
	if err = l.svcCtx.UserModel.Update(l.ctx, u); err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeUserUpdateFailed, err.Error())
	}
	return &types.ChangePasswordResponse{Success: true, Message: "ok"}, nil
}
