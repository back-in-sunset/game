package discovery

import (
	"context"

	"im/internal/contracts"
)

type Node = contracts.Node

type Registry interface {
	Register(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (Node, error)
	Close() error
}
