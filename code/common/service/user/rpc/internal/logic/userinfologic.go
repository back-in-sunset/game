package logic

import (
	"context"
	"strconv"
	"user/model"
	"user/rpc/internal/svc"
	"user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserInfoLogic) UserInfo(in *user.UserInfoRequest) (*user.UserInfoResponse, error) {
	md, ok := metadata.FromIncomingContext(l.ctx)
	if !ok {
		return nil, status.Error(100, "metadata not found")
	}
	// 从metadata中获取x-uid
	uidStrs := md.Get("x-uid")
	if len(uidStrs) == 0 {
		return nil, status.Error(100, "x-uid not found")
	}
	uid, err := strconv.ParseInt(uidStrs[0], 10, 64)
	if err != nil {
		return nil, status.Error(100, "x-uid invalid")
	}

	if in.ID <= 0 {
		return nil, status.Error(400, "用户ID不能为空")
	}
	if uid != in.ID {
		return nil, status.Error(403, "无权访问该用户信息")
	}

	// 查询用户是否存在
	res, err := l.svcCtx.UserModel.FindOne(l.ctx, in.ID)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(100, "用户不存在")
		}
		return nil, status.Error(500, err.Error())
	}

	return &user.UserInfoResponse{
		ID:     res.UserID,
		Name:   res.Name,
		Gender: res.Gender,
		Mobile: res.Mobile,
	}, nil
}
