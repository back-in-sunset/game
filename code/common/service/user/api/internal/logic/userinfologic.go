package logic

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/api/userclient"
	"user/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo() (resp *types.UserInfoResponse, err error) {
	// 查询用户是否存在
	uid, err := extractUID(l.ctx)
	if err != nil {
		return nil, err
	}

	md := metadata.Pairs("x-uid", strconv.FormatInt(uid, 10))
	l.ctx = metadata.NewOutgoingContext(l.ctx, md)

	// 查询用户信息
	res, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &userclient.UserInfoRequest{
		ID: uid,
	})
	if err != nil {
		return nil, err
	}

	out := &types.UserInfoResponse{
		ID:     res.ID,
		Name:   res.Name,
		Gender: res.Gender,
		Mobile: res.Mobile,
	}

	if l.svcCtx.UserProfileModel != nil {
		profile, pErr := l.svcCtx.UserProfileModel.FindOne(l.ctx, uid)
		if pErr == nil {
			out.Avatar = profile.Avatar
			out.Bio = profile.Bio
			if profile.Birthday.Valid {
				out.Birthday = profile.Birthday.Time.Format("2006-01-02")
			}
			out.Location = profile.Location
			if profile.Extra.Valid {
				out.Extra = profile.Extra.String
			}
		} else if pErr != nil && pErr != model.ErrNotFound {
			logx.WithContext(l.ctx).Errorf("load user profile failed, uid=%d err=%v", uid, pErr)
		}
	}
	return out, nil
}

func extractUID(ctx context.Context) (int64, error) {
	v := ctx.Value(types.UserIDKey)
	n, ok := v.(json.Number)
	if !ok {
		return 0, errx.New(http.StatusBadRequest, errx.CodeUIDInvalid, "uid invalid")
	}
	uid, err := n.Int64()
	if err != nil {
		return 0, errx.New(http.StatusBadRequest, errx.CodeUIDInvalid, "uid invalid")
	}
	return uid, nil
}
