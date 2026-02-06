package router

import (
	"easymeme/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(
	tokenHandler *handler.TokenHandler,
	tradeHandler *handler.TradeHandler,
	wsHub *handler.WebSocketHub,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.GET("/tokens", tokenHandler.GetTokens)
		api.GET("/tokens/:address", tokenHandler.GetToken)
		api.GET("/tokens/pending", tokenHandler.GetPendingTokens)
		api.POST("/tokens/:address/analysis", tokenHandler.PostTokenAnalysis)

		api.POST("/trades", tradeHandler.CreateTrade)
		api.GET("/trades", tradeHandler.GetTrades)
		api.PATCH("/trades/:txHash", tradeHandler.UpdateTradeStatus)
	}

	r.GET("/ws", wsHub.HandleWebSocket)

	return r
}
