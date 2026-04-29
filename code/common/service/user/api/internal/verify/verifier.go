package verify

import "context"

type CodeVerifier interface {
	SendCode(ctx context.Context, target string) error
	VerifyCode(ctx context.Context, target string, code string) error
}
