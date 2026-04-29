package verify

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type EmailConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	Subject    string
	TTL        time.Duration
	Cooldown   time.Duration
	DailyLimit int64
}

type emailVerifier struct {
	cfg     EmailConfig
	store   CodeStore
	limiter RateLimiter
}

func NewEmailVerifier(cfg EmailConfig, store CodeStore, limiter RateLimiter) CodeVerifier {
	if cfg.TTL <= 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.Subject == "" {
		cfg.Subject = "Verification Code"
	}
	return &emailVerifier{cfg: cfg, store: store, limiter: limiter}
}

func (v *emailVerifier) SendCode(ctx context.Context, target string) error {
	if v.limiter != nil {
		if err := v.limiter.AllowSend(ctx, "email_code", target, v.cfg.Cooldown, v.cfg.DailyLimit); err != nil {
			return err
		}
	}
	code, err := genCode6()
	if err != nil {
		return err
	}
	v.store.Set(target, code, v.cfg.TTL)
	if v.cfg.Host == "" || v.cfg.Port == 0 || v.cfg.From == "" {
		return fmt.Errorf("email verifier is not configured")
	}
	auth := smtp.PlainAuth("", v.cfg.Username, v.cfg.Password, v.cfg.Host)
	addr := fmt.Sprintf("%s:%d", v.cfg.Host, v.cfg.Port)
	msg := strings.Join([]string{
		"From: " + v.cfg.From,
		"To: " + target,
		"Subject: " + v.cfg.Subject,
		"",
		"Your verification code is: " + code,
	}, "\r\n")
	return smtp.SendMail(addr, auth, v.cfg.From, []string{target}, []byte(msg))
}

func (v *emailVerifier) VerifyCode(_ context.Context, target string, code string) error {
	if !v.store.VerifyAndDelete(target, code) {
		return fmt.Errorf("invalid verification code")
	}
	return nil
}
