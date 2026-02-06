package handler

import (
    "net/http"

    "easymeme/internal/repository"
    "easymeme/internal/service"

    "github.com/ethereum/go-ethereum/common"
    "github.com/gin-gonic/gin"
)

type TokenHandler struct {
    repo     *repository.Repository
    analyzer *service.Analyzer
}

func NewTokenHandler(repo *repository.Repository, analyzer *service.Analyzer) *TokenHandler {
    return &TokenHandler{repo: repo, analyzer: analyzer}
}

func (h *TokenHandler) GetTokens(c *gin.Context) {
    tokens, err := h.repo.GetLatestTokens(c.Request.Context(), 50)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": tokens})
}

func (h *TokenHandler) GetToken(c *gin.Context) {
    address := c.Param("address")
    token, err := h.repo.GetTokenByAddress(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": token})
}

func (h *TokenHandler) AnalyzeToken(c *gin.Context) {
    address := c.Param("address")
    if !common.IsHexAddress(address) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
        return
    }

    token, err := h.repo.GetTokenByAddress(c.Request.Context(), address)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
        return
    }

    result := h.analyzer.Analyze(c.Request.Context(), common.HexToAddress(address))
    token.RiskScore = result.Score
    token.RiskLevel = string(result.Level)
    token.IsHoneypot = result.IsHoneypot
    token.BuyTax = result.BuyTax
    token.SellTax = result.SellTax

    if err := h.repo.UpdateToken(c.Request.Context(), token); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": token})
}
