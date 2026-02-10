package router

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

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
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key", "X-User-Id", "X-Timestamp", "X-Nonce", "X-Signature"},
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
		api.GET("/tokens/stats/golden-dog-score-distribution", tokenHandler.GetGoldenDogScoreDistribution)
		api.GET("/tokens/:address/price-series", tokenHandler.GetTokenPriceSeries)
		api.POST("/tokens/price-snapshots", apiKeyMiddleware(cfg.ApiKey), tokenHandler.UpsertTokenPriceSnapshot)
		api.POST("/tokens/:address/analysis", apiKeyMiddleware(cfg.ApiKey), tokenHandler.PostTokenAnalysis)

		api.POST("/trades", tradeHandler.CreateTrade)
		api.GET("/trades", tradeHandler.GetTrades)
		api.PATCH("/trades/:txHash", tradeHandler.UpdateTradeStatus)

		api.GET("/wallet/info", walletHandler.GetWalletInfo)
		api.GET("/ai-positions", walletHandler.GetAIPositions)
		walletAuth := chainMiddleware(
			apiKeyUserMiddleware(cfg.ApiKey, cfg.ApiUserID),
			hmacMiddleware(cfg.ApiHmacSecret),
		)
		api.POST("/wallet/create", walletAuth, walletHandler.CreateWallet)
		api.GET("/wallet/balance", walletAuth, walletHandler.GetWalletBalance)
		api.POST("/wallet/withdraw", walletAuth, walletHandler.Withdraw)
		api.POST("/wallet/execute-trade", walletAuth, walletHandler.ExecuteTrade)
		api.POST("/wallet/config", walletAuth, walletHandler.UpsertWalletConfig)

		api.GET("/ai-trades", aiTradeHandler.GetAITrades)
		api.POST("/ai-trades", walletAuth, aiTradeHandler.CreateAITrade)
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

func apiKeyUserMiddleware(expectedKey, expectedUser string) gin.HandlerFunc {
	if expectedKey == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if expectedUser == "" {
			c.Next()
			return
		}
		userID := resolveUserID(c)
		if userID == "" || userID != expectedUser {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "userId mismatch"})
			return
		}
		c.Next()
	}
}

func resolveUserID(c *gin.Context) string {
	if userID := c.GetHeader("X-User-Id"); userID != "" {
		return userID
	}
	if userID := c.Query("userId"); userID != "" {
		return userID
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if userID, ok := payload["userId"].(string); ok {
		return userID
	}
	return ""
}

func chainMiddleware(middlewares ...gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, mw := range middlewares {
			mw(c)
			if c.IsAborted() {
				return
			}
		}
	}
}

type nonceStore struct {
	mu      sync.Mutex
	entries map[string]time.Time
	ttl     time.Duration
}

func newNonceStore(ttl time.Duration) *nonceStore {
	return &nonceStore{
		entries: make(map[string]time.Time),
		ttl:     ttl,
	}
}

func (s *nonceStore) seenOrAdd(nonce string) bool {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.entries {
		if now.Sub(v) > s.ttl {
			delete(s.entries, k)
		}
	}
	if _, exists := s.entries[nonce]; exists {
		return true
	}
	s.entries[nonce] = now
	return false
}

var globalNonceStore = newNonceStore(10 * time.Minute)

func hmacMiddleware(secret string) gin.HandlerFunc {
	if secret == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		ts := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		sig := c.GetHeader("X-Signature")
		if ts == "" || nonce == "" || sig == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing signature headers"})
			return
		}
		if globalNonceStore.seenOrAdd(nonce) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "nonce replay"})
			return
		}
		tsInt, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid timestamp"})
			return
		}
		now := time.Now().Unix()
		if tsInt < now-300 || tsInt > now+300 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "timestamp out of range"})
			return
		}

		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		payload := buildSignaturePayload(c, ts, nonce, body)
		expected := hmacSHA256Hex([]byte(secret), payload)
		if !hmac.Equal([]byte(expected), []byte(sig)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
		c.Next()
	}
}

func buildSignaturePayload(c *gin.Context, ts, nonce string, body []byte) []byte {
	uri := c.Request.URL.RequestURI()
	method := c.Request.Method
	return []byte(method + "\n" + uri + "\n" + ts + "\n" + nonce + "\n" + string(body))
}

func hmacSHA256Hex(secret []byte, payload []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
