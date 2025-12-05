package svc

import (
	"log"
	"time"
	"user/rpc/internal/config"
	"user/utils/idx"
)

func newIdxGenerator(c config.Config) *idx.Snowflake {
	// 分配节点 NodeID
	nodeID, lease, cli, err := idx.AllocateNodeID(
		c.Etcd.Hosts,
		c.Identity.Timeout,
	)
	if err != nil {
		log.Fatalf("allocate node id error: %v", err)
	}

	// 自动 keepalive
	err = idx.KeepAlive(cli, lease)
	if err != nil {
		log.Fatalf("keepalive error: %v", err)
	}

	return idx.NewSnowflakeWithPool(nodeID, c.Identity.PoolSize, c.Identity.BatchSize, time.Now().UnixMilli)
}
