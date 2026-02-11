package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	basesdk "easyweb3/go-sdk"

	_ "easyweb3/apps/easymeme/docs"
	"easyweb3/apps/easymeme/internal/config"
	"easyweb3/apps/easymeme/internal/handler"
	"easyweb3/apps/easymeme/internal/repository"
	"easyweb3/apps/easymeme/internal/router"
	"easyweb3/apps/easymeme/internal/service"
	"easyweb3/apps/easymeme/pkg/ethereum"
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

	baseClient := basesdk.NewClient(cfg.BaseServiceURL, cfg.BaseServiceToken)
	scanner := service.NewScanner(ethClient, repo, wsHub, baseClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scanner.Start(ctx); err != nil {
		log.Printf("Scanner not started: %v", err)
	}

	tokenHandler := handler.NewTokenHandler(repo)
	tradeHandler := handler.NewTradeHandler(repo)
	walletHandler := handler.NewWalletHandler(repo, ethClient)
	aiTradeHandler := handler.NewAITradeHandler(repo)

	r := router.Setup(cfg, tokenHandler, tradeHandler, walletHandler, aiTradeHandler, wsHub, scanner)

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
