package logic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"platform/api/internal/svc"
	"platform/api/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

const demoTokenTTL = 3600 // 1 hour

type DemoTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDemoTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoTokenLogic {
	return &DemoTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoTokenLogic) DemoToken(req *types.DemoTokenReq) (resp *types.DemoTokenResp, err error) {
	now := time.Now().Unix()
	uid := int64(1) // fixed demo user

	// 1. IM JWT (RS256) — same format as user service
	imToken, err := l.signIMToken(uid, now)
	if err != nil {
		return nil, fmt.Errorf("sign IM token: %w", err)
	}

	// 2. LiveKit JWT (HS256)
	lkToken, err := l.signLiveKitToken(now)
	if err != nil {
		return nil, fmt.Errorf("sign LiveKit token: %w", err)
	}

	// Combine tokens: header.payload for each, separated by |
	// Frontend splits and uses individually
	combined := fmt.Sprintf("%s|%s", imToken, lkToken)

	return &types.DemoTokenResp{
		AccessToken: combined,
		ExpiresIn:   demoTokenTTL,
		LiveKitUrl:  "http://localhost:7880",
	}, nil
}

func (l *DemoTokenLogic) signIMToken(uid int64, now int64) (string, error) {
	claims := jwt.MapClaims{
		"exp": now + demoTokenTTL,
		"iat": now,
		"uid": uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(l.svcCtx.PrivateKey)
}

func (l *DemoTokenLogic) signLiveKitToken(now int64) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := map[string]any{
		"iss": l.svcCtx.Config.LiveKit.ApiKey,
		"sub": "demo-user",
		"nbf": now,
		"exp": now + demoTokenTTL,
		"video": map[string]any{
			"room":         "demo-room-*",
			"roomJoin":     true,
			"canPublish":   true,
			"canSubscribe": true,
		},
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64url(headerJSON)
	payloadB64 := base64url(payloadJSON)

	signingInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(l.svcCtx.Config.LiveKit.ApiSecret))
	mac.Write([]byte(signingInput))
	sig := base64url(mac.Sum(nil))

	return signingInput + "." + sig, nil
}

func base64url(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}
