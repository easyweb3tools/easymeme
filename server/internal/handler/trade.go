package handler

import (
    "net/http"

    "easymeme/internal/model"
    "easymeme/internal/repository"

    "github.com/gin-gonic/gin"
)

type TradeHandler struct {
    repo *repository.Repository
}

func NewTradeHandler(repo *repository.Repository) *TradeHandler {
    return &TradeHandler{repo: repo}
}

type CreateTradeRequest struct {
    UserAddress  string `json:"user_address"`
    TokenAddress string `json:"token_address"`
    TokenSymbol  string `json:"token_symbol"`
    Type         string `json:"type"`
    AmountIn     string `json:"amount_in"`
    AmountOut    string `json:"amount_out"`
    TxHash       string `json:"tx_hash"`
    Status       string `json:"status"`
    GasUsed      string `json:"gas_used"`
}

func (h *TradeHandler) CreateTrade(c *gin.Context) {
    var req CreateTradeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }

    trade := &model.Trade{
        UserAddress:  req.UserAddress,
        TokenAddress: req.TokenAddress,
        TokenSymbol:  req.TokenSymbol,
        Type:         req.Type,
        TxHash:       req.TxHash,
        Status:       req.Status,
    }

    if err := h.repo.CreateTrade(c.Request.Context(), trade); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": trade})
}

func (h *TradeHandler) GetTrades(c *gin.Context) {
    user := c.Query("user")
    if user == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "user is required"})
        return
    }

    trades, err := h.repo.GetTradesByUser(c.Request.Context(), user, 50)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": trades})
}

func (h *TradeHandler) UpdateTradeStatus(c *gin.Context) {
    txHash := c.Param("txHash")

    var req struct {
        Status string `json:"status"`
    }
    if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }

    if err := h.repo.UpdateTradeStatus(c.Request.Context(), txHash, req.Status); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": gin.H{"tx_hash": txHash, "status": req.Status}})
}
