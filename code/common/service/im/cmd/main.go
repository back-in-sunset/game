package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"im/internal/config"
	"im/internal/service"
)

var configFile = flag.String("f", "etc/im.yaml", "config file")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app, err := service.NewApp(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("starting im service, ws=%s tcp=%s\n", cfg.Listen.WebSocket, cfg.Listen.TCP)
	if err := app.Start(ctx); err != nil {
		log.Fatalf("start app: %v", err)
	}
	<-ctx.Done()
	app.Stop(context.Background())
}
