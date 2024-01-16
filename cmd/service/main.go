package main

import (
	"context"
	"fmt"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/cache"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/clock"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/config"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/storage"
	"github.com/kkucherenkov/faraway-challenge/internal/service"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("start server")

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		logger.Error("config load error", "error", err)
		return
	}

	clockSvc := clock.SystemClock{}
	cacheInst := cache.NewLocalCache(clockSvc) // InitRedisCache(ctx, configInst.CacheHost, configInst.CachePort)
	if err != nil {
		logger.Error("cache init error", "error", err)
		return
	}

	storeSvc := storage.CreateStorage(cfg.Service.DataFile)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", cfg)
	ctx = context.WithValue(ctx, "clock", clockSvc)
	ctx = context.WithValue(ctx, "cache", cacheInst)
	ctx = context.WithValue(ctx, "storage", storeSvc)
	ctx = context.WithValue(ctx, "logger", logger)

	// run server
	serverAddress := fmt.Sprintf("%s:%s", cfg.Service.Host, cfg.Service.Port)
	err = service.Run(ctx, serverAddress)
	if err != nil {
		logger.Error("server error", "error", err)
	}
}
