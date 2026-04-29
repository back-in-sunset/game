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

type SendMobileCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMobileCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMobileCodeLogic {
	return &SendMobileCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMobileCodeLogic) SendMobileCode(req *types.SendMobileCodeRequest) (*types.SendCodeResponse, error) {
	req.Mobile = strings.TrimSpace(req.Mobile)
	if req.Mobile == "" {
		return nil, errx.New(http.StatusBadRequest, errx.CodeMobileRequired, "mobile is required")
	}
	if l.svcCtx.SmsVerifier == nil {
		return nil, errx.New(http.StatusInternalServerError, errx.CodeSMSVerifierUnavailable, "sms verifier unavailable")
	}
	if err := l.svcCtx.SmsVerifier.SendCode(l.ctx, req.Mobile); err != nil {
		return nil, errx.New(http.StatusTooManyRequests, errx.CodeRateLimited, err.Error())
	}
	return &types.SendCodeResponse{Success: true, Message: "ok"}, nil
}
