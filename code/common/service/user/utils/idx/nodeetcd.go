package idx

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// SnowflakePrefix 雪花算法节点ID前缀
	SnowflakePrefix = "/snowflake/nodes/"
	// MaxNodeID 最大节点ID
	MaxNodeID = 1024 // 10 bits
)

// AllocateNodeID 分配节点ID
func AllocateNodeID(endpoints []string, timeoutSec int) (int64, clientv3.LeaseID, *clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(timeoutSec) * time.Second,
	})
	if err != nil {
		return 0, 0, nil, err
	}

	leaseResp, err := cli.Grant(context.Background(), 10)
	if err != nil {
		return 0, 0, nil, err
	}
	leaseID := leaseResp.ID

	for id := int64(0); id < MaxNodeID; id++ {
		key := fmt.Sprintf("%s%d", SnowflakePrefix, id)

		txnResp, err := cli.Txn(context.Background()).
			If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
			Then(clientv3.OpPut(key, "active", clientv3.WithLease(leaseID))).
			Commit()

		if err != nil {
			return 0, 0, nil, err
		}

		if txnResp.Succeeded {
			return id, leaseID, cli, nil
		}
	}

	return 0, 0, nil, fmt.Errorf("no node id available")
}

// KeepAlive 保持租约_alive
func KeepAlive(cli *clientv3.Client, lease clientv3.LeaseID) error {
	ch, err := cli.KeepAlive(context.Background(), lease)
	if err != nil {
		return err
	}
	go func() {
		for range ch {
			// ignore
		}
	}()
	return nil
}
