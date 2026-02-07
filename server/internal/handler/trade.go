package handler

import (
	"log"
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

type TradeResponseEnvelope struct {
	Data model.Trade `json:"data"`
}

// CreateTrade godoc
// @Summary Create trade
// @Description Create trade record
// @Tags trades
// @Param payload body CreateTradeRequest true "Trade payload"
// @Success 200 {object} TradeResponseEnvelope
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/trades [post]
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
		log.Printf("create trade: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": trade})
}

type TradeListResponseEnvelope struct {
	Data []model.Trade `json:"data"`
}

// GetTrades godoc
// @Summary Get trades
// @Description Get trades by user address
// @Tags trades
// @Param user query string true "User address"
// @Success 200 {object} TradeListResponseEnvelope
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/trades [get]
func (h *TradeHandler) GetTrades(c *gin.Context) {
	user := c.Query("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user is required"})
		return
	}

	trades, err := h.repo.GetTradesByUser(c.Request.Context(), user, 50)
	if err != nil {
		log.Printf("get trades: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": trades})
}

type TradeStatusUpdateRequest struct {
	Status string `json:"status"`
}

type TradeStatusUpdateResponseEnvelope struct {
	Data map[string]string `json:"data"`
}

// UpdateTradeStatus godoc
// @Summary Update trade status
// @Description Update trade status by tx hash
// @Tags trades
// @Param txHash path string true "Transaction hash"
// @Param payload body TradeStatusUpdateRequest true "Status payload"
// @Success 200 {object} TradeStatusUpdateResponseEnvelope
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/trades/{txHash} [patch]
func (h *TradeHandler) UpdateTradeStatus(c *gin.Context) {
	txHash := c.Param("txHash")

	var req TradeStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.repo.UpdateTradeStatus(c.Request.Context(), txHash, req.Status); err != nil {
		log.Printf("update trade status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"tx_hash": txHash, "status": req.Status}})
}
