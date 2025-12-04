package svc

import (
	"log"
	"time"
	"user/model"
	"user/rpc/internal/config"
	"user/utils/idx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	UserModel model.UserModel

	IdxGen *idx.Snowflake
}

// NewServiceContext 创建 ServiceContext
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)

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

	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn, c.CacheRedis),
		IdxGen:    idx.NewSnowflakeWithPool(nodeID, c.Identity.PoolSize, c.Identity.BatchSize, time.Now().UnixMilli),
	}
}
