package scope

import (
	"im/internal/config"
	"im/internal/domain"
)

type Resolver struct {
	defaultEnvironment string
}

func NewResolver(cfg config.ScopeDefaults) *Resolver {
	return &Resolver{defaultEnvironment: cfg.DefaultEnvironment}
}

func (r *Resolver) Resolve(imDomain domain.IMDomain, scope domain.Scope) (domain.Scope, error) {
	scope = scope.Normalize()
	if imDomain == domain.DomainTenant && scope.Environment == "" && r.defaultEnvironment != "" {
		scope.Environment = r.defaultEnvironment
	}
	return scope, scope.Validate(imDomain)
}
