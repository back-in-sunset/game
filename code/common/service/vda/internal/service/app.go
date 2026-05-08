package service

import (
	"context"
	"fmt"
	"net"

	"vda/internal/callmanager"
	"vda/internal/config"
	"vda/internal/livekit"
	"vda/internal/roommanager"
	"vda/internal/storage"
	rpc "vda/rpc/vdaclient"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// App 封装 VDA 服务的所有依赖和生命周期。
type App struct {
	cfg        config.Config
	livekit    *livekit.Client
	store      *storage.RedisStore
	grpcServer *grpc.Server
	rdb        *redis.Client
	callMgr    *callmanager.Manager
	roomMgr    *roommanager.Manager
}

// NewApp 创建 VDA 应用实例，初始化 Redis、LiveKit 客户端和业务管理器。
func NewApp(cfg config.Config) (*App, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	lk := livekit.New(cfg.LiveKit.Host, cfg.LiveKit.APIKey, cfg.LiveKit.APISecret)
	store := storage.NewRedisStore(rdb, cfg.Redis.KeyPrefix)

	return &App{
		cfg:     cfg,
		livekit: lk,
		store:   store,
		rdb:     rdb,
		callMgr: callmanager.New(lk, store, rdb),
		roomMgr: roommanager.New(lk, store, rdb),
	}, nil
}

// Start 启动 gRPC 服务器并注册 VDA 服务。
func (a *App) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", a.cfg.Listen.RPC)
	if err != nil {
		return fmt.Errorf("listen rpc %s: %w", a.cfg.Listen.RPC, err)
	}

	srv := grpc.NewServer()
	a.grpcServer = srv

	rpc.RegisterVDAServer(srv, NewVDAServer(a.callMgr, a.roomMgr, a.livekit.Host()))
	reflection.Register(srv)

	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return nil
}

// Stop 优雅关闭 gRPC 服务器并释放 Redis 连接。
func (a *App) Stop(ctx context.Context) {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.rdb != nil {
		a.rdb.Close()
	}
}
