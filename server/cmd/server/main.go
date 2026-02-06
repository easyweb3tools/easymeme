package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "easymeme/internal/config"
    "easymeme/internal/handler"
    "easymeme/internal/repository"
    "easymeme/internal/router"
    "easymeme/internal/service"
    "easymeme/pkg/ethereum"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    repo, err := repository.New(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect database: %v", err)
    }
    log.Println("Database connected")

    ethClient, err := ethereum.NewClient(cfg.BscRpcHTTP, cfg.BscRpcWS)
    if err != nil {
        log.Fatalf("Failed to connect BSC: %v", err)
    }
    defer ethClient.Close()
    log.Println("BSC RPC connected")

    wsHub := handler.NewWebSocketHub()
    go wsHub.Run()

    analyzer := service.NewAnalyzer(ethClient)
    scanner := service.NewScanner(ethClient, repo, analyzer, wsHub)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := scanner.Start(ctx); err != nil {
        log.Printf("Scanner not started: %v", err)
    }

    tokenHandler := handler.NewTokenHandler(repo, analyzer)
    tradeHandler := handler.NewTradeHandler(repo)

    r := router.Setup(tokenHandler, tradeHandler, wsHub)

    go func() {
        log.Printf("Server starting on port %s", cfg.Port)
        if err := r.Run(":" + cfg.Port); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down...")
    cancel()
}
