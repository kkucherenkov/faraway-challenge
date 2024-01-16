package main

import (
	"context"
	"fmt"
	"github.com/kkucherenkov/faraway-challenge/internal/client"
	"github.com/kkucherenkov/faraway-challenge/internal/pkg/config"
	"log/slog"
	"os"
	"sync"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("start client")

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		logger.Error("config load error", "error", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", cfg)
	ctx = context.WithValue(ctx, "logger", logger)

	serverAddress := fmt.Sprintf("%s:%s", cfg.Service.Host, cfg.Service.Port)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := client.RunSync(ctx, serverAddress)
		if err != nil {
			if err != nil {
				logger.Error("sync client error", "error", err)
			}
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		err := client.RunAsync(ctx)
		if err != nil {
			if err != nil {
				logger.Error("async client error", "error", err)
			}
		}
		wg.Done()
	}()
	wg.Wait()
}
