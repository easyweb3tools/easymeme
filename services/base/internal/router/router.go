package router

import (
	"easyweb3/base/internal/config"
	"easyweb3/base/internal/handler"
	"easyweb3/base/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Setup(
	cfg *config.Config,
	tokenSecurityHandler *handler.TokenSecurityHandler,
	marketDataHandler *handler.MarketDataHandler,
	chainDataHandler *handler.ChainDataHandler,
	walletHandler *handler.WalletHandler,
	notificationHandler *handler.NotificationHandler,
) *gin.Engine {
	r := gin.Default()

	origins := cfg.CorsAllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "base"})
	})

	v1 := r.Group("/api/v1")
	v1.Use(middleware.ServiceAuthMiddleware(cfg.ServiceTokens))
	{
		v1.GET("/tokens/:chain/:address/security", tokenSecurityHandler.GetTokenSecurity)
		v1.GET("/tokens/:chain/:address/holders", chainDataHandler.GetHolderDistribution)
		v1.GET("/tokens/:chain/:address/creator", chainDataHandler.GetCreatorHistory)
		v1.GET("/market/:chain/pairs/:pairAddress", marketDataHandler.GetPairData)

		v1.POST("/wallets", walletHandler.CreateWallet)
		v1.GET("/wallets/:walletId/balance", walletHandler.GetWalletBalance)
		v1.POST("/wallets/:walletId/execute", walletHandler.ExecuteTrade)

		v1.POST("/notifications/send", notificationHandler.SendNotification)
	}

	return r
}
