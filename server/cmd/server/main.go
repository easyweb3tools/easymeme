package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "easymeme/docs"
	"easymeme/internal/config"
	"easymeme/internal/handler"
	"easymeme/internal/repository"
	"easymeme/internal/router"
	"easymeme/internal/service"
	"easymeme/pkg/ethereum"
)

//go:generate swag init -g cmd/server/main.go -o docs

// @title EasyMeme API
// @version 0.1
// @description EasyMeme server API for token analysis and discovery.
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
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

	scanner := service.NewScanner(ethClient, repo, wsHub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scanner.Start(ctx); err != nil {
		log.Printf("Scanner not started: %v", err)
	}

	tokenHandler := handler.NewTokenHandler(repo)
	tradeHandler := handler.NewTradeHandler(repo)
	walletHandler := handler.NewWalletHandler(repo, ethClient)
	aiTradeHandler := handler.NewAITradeHandler(repo)

	r := router.Setup(cfg, tokenHandler, tradeHandler, walletHandler, aiTradeHandler, wsHub)

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
