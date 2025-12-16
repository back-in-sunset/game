package svc

import (
	"log"
	"time"
	"user/model"
	"user/rpc/internal/config"

	"github.com/gocql/gocql"
)

func newCQLUserModel(c config.Config) model.UserModel {
	cluster := gocql.NewCluster(
		c.ScyllaDB.Hosts...,
	)

	cluster.Keyspace = c.ScyllaDB.Keyspace
	cluster.Consistency = gocql.Consistency(c.ScyllaDB.Consistency)
	cluster.Timeout = time.Duration(c.ScyllaDB.Timeout) * time.Second
	cluster.NumConns = c.ScyllaDB.NumConns // 高 QPS 建议 >10
	cluster.ReconnectInterval = time.Duration(c.ScyllaDB.ReconnectInterval) * time.Second

	cluster.CreateSession()
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("create session error: %v", err)
	}

	return model.NewCQLUserModel(session, c.CacheRedis)
}
