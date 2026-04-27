package svc

import "context"

type Scope struct {
	TenantID    string
	ProjectID   string
	Environment string
}

type scopeContextKey struct{}

func WithScope(ctx context.Context, scope Scope) context.Context {
	return context.WithValue(ctx, scopeContextKey{}, scope)
}

func ScopeFromContext(ctx context.Context) (Scope, bool) {
	scope, ok := ctx.Value(scopeContextKey{}).(Scope)
	return scope, ok
}

func ResolveScope(ctx context.Context, fallback Scope) Scope {
	if scope, ok := ScopeFromContext(ctx); ok {
		return scope
	}
	return fallback
}
