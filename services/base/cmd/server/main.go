package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"easyweb3/base/internal/config"
	"easyweb3/base/internal/handler"
	"easyweb3/base/internal/repository"
	"easyweb3/base/internal/router"
	"easyweb3/base/internal/service"
	"easyweb3/base/pkg/ethereum"
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

	cache, err := service.NewCacheService(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect Redis: %v", err)
	}
	log.Println("Redis connected")

	ethClient, err := ethereum.NewClient(cfg.BscRpcHTTP, cfg.BscRpcWS)
	if err != nil {
		log.Fatalf("Failed to connect BSC: %v", err)
	}
	defer ethClient.Close()
	log.Println("BSC RPC connected")

	goplusClient := service.NewGoPlusClient()
	dexClient := service.NewDEXScreenerClient()
	scanClient := service.NewBscScanClient(cfg.BscScanAPIKey)
	telegramClient := service.NewTelegramClient(cfg.TelegramBotToken)

	tokenSecurityHandler := handler.NewTokenSecurityHandler(goplusClient, cache)
	marketDataHandler := handler.NewMarketDataHandler(dexClient, cache)
	chainDataHandler := handler.NewChainDataHandler(scanClient, cache)
	walletHandler := handler.NewWalletHandler(repo, ethClient, cfg.WalletMasterKey)
	notificationHandler := handler.NewNotificationHandler(telegramClient, repo)

	r := router.Setup(cfg, tokenSecurityHandler, marketDataHandler, chainDataHandler, walletHandler, notificationHandler)

	go func() {
		log.Printf("Base service starting on port %s", cfg.Port)
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
