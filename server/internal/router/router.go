package router

import (
	"net/http"

	"easymeme/internal/config"
	"easymeme/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(
	cfg *config.Config,
	tokenHandler *handler.TokenHandler,
	tradeHandler *handler.TradeHandler,
	walletHandler *handler.WalletHandler,
	aiTradeHandler *handler.AITradeHandler,
	wsHub *handler.WebSocketHub,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CorsAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.GET("/tokens", tokenHandler.GetTokens)
		api.GET("/tokens/:address", tokenHandler.GetToken)
		api.GET("/tokens/:address/detail", tokenHandler.GetTokenDetail)
		api.GET("/tokens/pending", tokenHandler.GetPendingTokens)
		api.GET("/tokens/analyzed", tokenHandler.GetAnalyzedTokens)
		api.GET("/tokens/golden-dogs", tokenHandler.GetGoldenDogs)
		api.POST("/tokens/:address/analysis", apiKeyMiddleware(cfg.ApiKey), tokenHandler.PostTokenAnalysis)

		api.POST("/trades", tradeHandler.CreateTrade)
		api.GET("/trades", tradeHandler.GetTrades)
		api.PATCH("/trades/:txHash", tradeHandler.UpdateTradeStatus)

		api.POST("/wallet/create", walletHandler.CreateWallet)
		api.GET("/wallet/balance", walletHandler.GetWalletBalance)
		api.POST("/wallet/withdraw", walletHandler.Withdraw)
		api.POST("/wallet/execute-trade", walletHandler.ExecuteTrade)
		api.POST("/wallet/config", walletHandler.UpsertWalletConfig)

		api.GET("/ai-trades", aiTradeHandler.GetAITrades)
		api.POST("/ai-trades", aiTradeHandler.CreateAITrade)
		api.GET("/ai-trades/stats", aiTradeHandler.GetAITradeStats)
	}

	r.GET("/ws", wsHub.HandleWebSocket)

	return r
}

func apiKeyMiddleware(expected string) gin.HandlerFunc {
	if expected == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
