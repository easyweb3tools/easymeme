package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"easyweb3/base/internal/service"

	"github.com/gin-gonic/gin"
)

type MarketDataHandler struct {
	dexScreener *service.DEXScreenerClient
	cache       *service.CacheService
}

func NewMarketDataHandler(dex *service.DEXScreenerClient, cache *service.CacheService) *MarketDataHandler {
	return &MarketDataHandler{dexScreener: dex, cache: cache}
}

func (h *MarketDataHandler) GetPairData(c *gin.Context) {
	chain := strings.ToLower(strings.TrimSpace(c.Param("chain")))
	pairAddress := strings.TrimSpace(c.Param("pairAddress"))
	cacheKey := fmt.Sprintf("pair_data:%s:%s", chain, strings.ToLower(pairAddress))

	var cached map[string]interface{}
	if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
		c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
		return
	}

	data, err := h.dexScreener.GetPairData(c.Request.Context(), chain, pairAddress)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("dexscreener: %s", err.Error())})
		return
	}

	_ = h.cache.Set(c.Request.Context(), cacheKey, data, time.Minute)
	c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}
