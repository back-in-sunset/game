package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"
	"user/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileRequest) (*types.UpdateProfileResponse, error) {
	uid, err := l.ctx.Value(types.UserIDKey).(json.Number).Int64()
	if err != nil {
		return nil, errx.New(http.StatusBadRequest, errx.CodeUIDInvalid, "uid invalid")
	}

	req.Avatar = strings.TrimSpace(req.Avatar)
	req.Bio = strings.TrimSpace(req.Bio)
	req.Location = strings.TrimSpace(req.Location)
	req.Extra = strings.TrimSpace(req.Extra)

	if len(req.Avatar) > 1024 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeAvatarTooLong, "avatar length must be <= 1024")
	}
	if len(req.Bio) > 1024 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeBioTooLong, "bio length must be <= 1024")
	}
	if len(req.Location) > 255 {
		return nil, errx.New(http.StatusBadRequest, errx.CodeLocationTooLong, "location length must be <= 255")
	}

	var birthday sql.NullTime
	if req.Birthday != "" {
		t, parseErr := time.Parse("2006-01-02", req.Birthday)
		if parseErr != nil {
			return nil, errx.New(http.StatusBadRequest, errx.CodeBirthdayFormatInvalid, "birthday must use YYYY-MM-DD")
		}
		birthday = sql.NullTime{Time: t, Valid: true}
	}

	if l.svcCtx.UserProfileModel == nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeProfileModelUnavailable, "profile model unavailable")
	}

	err = l.svcCtx.UserProfileModel.Upsert(l.ctx, &model.UserProfile{
		UserID:   uid,
		Avatar:   req.Avatar,
		Bio:      req.Bio,
		Birthday: birthday,
		Location: req.Location,
		Extra:    sql.NullString{String: req.Extra, Valid: req.Extra != ""},
	})
	if err != nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeProfileUpsertFailed, err.Error())
	}

	return &types.UpdateProfileResponse{
		Success: true,
		Message: "ok",
	}, nil
}
