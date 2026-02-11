package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"easyweb3/base/internal/service"

	"github.com/gin-gonic/gin"
)

type TokenSecurityHandler struct {
	goPlus *service.GoPlusClient
	cache  *service.CacheService
}

func NewTokenSecurityHandler(goPlus *service.GoPlusClient, cache *service.CacheService) *TokenSecurityHandler {
	return &TokenSecurityHandler{goPlus: goPlus, cache: cache}
}

func (h *TokenSecurityHandler) GetTokenSecurity(c *gin.Context) {
	chain := strings.ToLower(strings.TrimSpace(c.Param("chain")))
	address := strings.TrimSpace(c.Param("address"))
	cacheKey := fmt.Sprintf("token_security:%s:%s", chain, strings.ToLower(address))

	var cached service.GoPlusSecurityData
	if err := h.cache.Get(c.Request.Context(), cacheKey, &cached); err == nil {
		c.JSON(http.StatusOK, gin.H{"data": cached, "cached": true})
		return
	}

	data, err := h.goPlus.GetTokenSecurity(c.Request.Context(), chain, address)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("goplus: %s", err.Error())})
		return
	}

	_ = h.cache.Set(c.Request.Context(), cacheKey, data, time.Hour)
	c.JSON(http.StatusOK, gin.H{"data": data, "cached": false})
}
