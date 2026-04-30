package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistry struct {
	client *clientv3.Client
	prefix string
	ttl    int64

	leaseID clientv3.LeaseID
	cancel  context.CancelFunc
}

func NewEtcdRegistry(endpoints []string, prefix string, ttlSeconds int64) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if ttlSeconds <= 0 {
		ttlSeconds = 10
	}

	return &EtcdRegistry{
		client: client,
		prefix: prefix,
		ttl:    ttlSeconds,
	}, nil
}

func (r *EtcdRegistry) Register(ctx context.Context, node Node) error {
	lease, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(node)
	if err != nil {
		return err
	}

	key := path.Join(r.prefix, node.ID)
	if _, err := r.client.Put(ctx, key, string(payload), clientv3.WithLease(lease.ID)); err != nil {
		return err
	}

	keepAliveCtx, cancel := context.WithCancel(context.Background())
	ch, err := r.client.KeepAlive(keepAliveCtx, lease.ID)
	if err != nil {
		cancel()
		return fmt.Errorf("keepalive: %w", err)
	}
	r.leaseID = lease.ID
	r.cancel = cancel

	go func() {
		for range ch {
		}
	}()

	return nil
}

func (r *EtcdRegistry) GetNode(ctx context.Context, id string) (Node, error) {
	key := path.Join(r.prefix, id)
	resp, err := r.client.Get(ctx, key)
	if err != nil {
		return Node{}, err
	}
	if len(resp.Kvs) == 0 {
		return Node{}, fmt.Errorf("node %s not found", id)
	}
	var node Node
	if err := json.Unmarshal(resp.Kvs[0].Value, &node); err != nil {
		return Node{}, err
	}
	return node, nil
}

func (r *EtcdRegistry) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
