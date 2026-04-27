package logic

import (
	"context"
	"strings"

	"user/model"
	"user/rpc/internal/svc"
	"user/rpc/user"
	"user/utils/cryptx"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Mobile = strings.TrimSpace(in.Mobile)
	if in.Name == "" {
		return nil, status.Error(400, "用户名不能为空")
	}
	if in.Mobile == "" {
		return nil, status.Error(400, "手机号不能为空")
	}
	if in.Password == "" {
		return nil, status.Error(400, "密码不能为空")
	}

	// 判断手机号是否已经注册
	_, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err == nil {
		return nil, status.Error(100, "该用户已存在")
	}

	if err == model.ErrNotFound {
		newUser := model.User{
			UserID:   l.svcCtx.IdxGen.Next(),
			Name:     in.Name,
			Gender:   in.Gender,
			Mobile:   in.Mobile,
			Password: cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, in.Password),
		}

		res, err := l.svcCtx.UserModel.Insert(l.ctx, &newUser)
		if err != nil {
			return nil, status.Error(500, err.Error())
		}

		if _, err = res.RowsAffected(); err != nil {
			return nil, status.Error(500, err.Error())
		}

		return &user.RegisterResponse{
			ID:     newUser.UserID,
			Name:   newUser.Name,
			Gender: newUser.Gender,
			Mobile: newUser.Mobile,
		}, nil

	}

	return nil, status.Error(500, err.Error())
}
