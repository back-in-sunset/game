package logic

import (
	"context"
	"net/http"
	"strings"

	"user/api/internal/errx"
	"user/api/internal/svc"
	"user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendEmailCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendEmailCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendEmailCodeLogic {
	return &SendEmailCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendEmailCodeLogic) SendEmailCode(req *types.SendEmailCodeRequest) (*types.SendCodeResponse, error) {
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeEmailRequired, "email is required")
	}
	if l.svcCtx.EmailVerifier == nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeEmailVerifierUnavailable, "email verifier unavailable")
	}
	if err := l.svcCtx.EmailVerifier.SendCode(l.ctx, req.Email); err != nil {
		return nil, errx.New(http.StatusTooManyRequests, errx.CodeRateLimited, err.Error())
	}
	return &types.SendCodeResponse{Success: true, Message: "ok"}, nil
}
