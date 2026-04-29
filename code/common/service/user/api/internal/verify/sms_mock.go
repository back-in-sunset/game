package verify

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type SmsMockConfig struct {
	FixedCode  string
	TTL        time.Duration
	Cooldown   time.Duration
	DailyLimit int64
}

type smsMockVerifier struct {
	cfg     SmsMockConfig
	store   CodeStore
	limiter RateLimiter
}

func NewSmsMockVerifier(cfg SmsMockConfig, store CodeStore, limiter RateLimiter) CodeVerifier {
	if cfg.TTL <= 0 {
		cfg.TTL = 5 * time.Minute
	}
	return &smsMockVerifier{cfg: cfg, store: store, limiter: limiter}
}

func (v *smsMockVerifier) SendCode(ctx context.Context, target string) error {
	if v.limiter != nil {
		if err := v.limiter.AllowSend(ctx, "sms_code", target, v.cfg.Cooldown, v.cfg.DailyLimit); err != nil {
			return err
		}
	}
	code := v.cfg.FixedCode
	if code == "" {
		var err error
		code, err = genCode6()
		if err != nil {
			return err
		}
	}
	v.store.Set(target, code, v.cfg.TTL)
	logx.Infof("[sms-mock] send code target=%s code=%s", target, code)
	return nil
}

func (v *smsMockVerifier) VerifyCode(_ context.Context, target string, code string) error {
	if !v.store.VerifyAndDelete(target, code) {
		return fmt.Errorf("invalid verification code")
	}
	return nil
}
