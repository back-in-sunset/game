package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"vda/internal/config"
	"vda/internal/service"
)

// VDA (Voice Data Audio) — 语音业务信令服务。
// 负责 1v1 通话和语音房间的信令管理，签发 LiveKit token。
var configFile = flag.String("f", "etc/vda.yaml", "config file")

func main() {
	flag.Parse()

	// 加载配置。
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 构建应用（Redis、LiveKit 客户端、gRPC 服务）。
	app, err := service.NewApp(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	// 监听 SIGINT/SIGTERM 实现优雅关闭。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("starting vda service, rpc=%s\n", cfg.Listen.RPC)
	if err := app.Start(ctx); err != nil {
		log.Fatalf("start app: %v", err)
	}
	<-ctx.Done()
	app.Stop(context.Background())
}
