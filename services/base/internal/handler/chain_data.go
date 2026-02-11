package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"easyweb3/base/internal/service"

	"github.com/gin-gonic/gin"
)

type ChainDataHandler struct {
	scan  *service.BscScanClient
	cache *service.CacheService
}

func NewChainDataHandler(scan *service.BscScanClient, cache *service.CacheService) *ChainDataHandler {
	return &ChainDataHandler{scan: scan, cache: cache}
}

func (h *ChainDataHandler) GetHolderDistribution(c *gin.Context) {
	chain := strings.ToLower(strings.TrimSpace(c.Param("chain")))
	address := strings.TrimSpace(c.Param("address"))
	cacheKey := fmt.Sprintf("holders:%s:%s", chain, strings.ToLower(address))

	var cached service.HolderDistribution
	if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
		c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
		return
	}

	data, err := h.scan.FetchHolderDistribution(c.Request.Context(), chain, address)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("scan: %s", err.Error())})
		return
	}

	_ = h.cache.Set(c.Request.Context(), cacheKey, data, 30*time.Minute)
	c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}

func (h *ChainDataHandler) GetCreatorHistory(c *gin.Context) {
	chain := strings.ToLower(strings.TrimSpace(c.Param("chain")))
	address := strings.TrimSpace(c.Param("address"))
	cacheKey := fmt.Sprintf("creator:%s:%s", chain, strings.ToLower(address))

	var cached service.CreatorHistory
	if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
		c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
		return
	}

	data, err := h.scan.FetchCreatorHistory(c.Request.Context(), chain, address)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("scan: %s", err.Error())})
		return
	}

	_ = h.cache.Set(c.Request.Context(), cacheKey, data, time.Hour)
	c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}
