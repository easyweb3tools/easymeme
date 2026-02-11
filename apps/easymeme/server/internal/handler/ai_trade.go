package handler

import (
	"log"
	"net/http"
	"strconv"

	"easyweb3/apps/easymeme/internal/model"
	"easyweb3/apps/easymeme/internal/repository"

	"github.com/gin-gonic/gin"
)

type AITradeHandler struct {
	repo *repository.Repository
}

func NewAITradeHandler(repo *repository.Repository) *AITradeHandler {
	return &AITradeHandler{repo: repo}
}

type AITradeListResponseEnvelope struct {
	Data []model.AITrade `json:"data"`
}

type CreateAITradeRequest struct {
	UserID         string  `json:"userId"`
	TokenAddress   string  `json:"tokenAddress"`
	TokenSymbol    string  `json:"tokenSymbol"`
	Type           string  `json:"type"`
	AmountIn       string  `json:"amountIn"`
	AmountOut      string  `json:"amountOut"`
	TxHash         string  `json:"txHash"`
	GoldenDogScore int     `json:"goldenDogScore"`
	DecisionReason string  `json:"decisionReason"`
	StrategyUsed   string  `json:"strategyUsed"`
	CurrentValue   string  `json:"currentValue"`
	ProfitLoss     float64 `json:"profitLoss"`
}

type AITradeResponseEnvelope struct {
	Data model.AITrade `json:"data"`
}

// CreateAITrade godoc
// @Summary Create AI trade
// @Description Create AI trade record
// @Tags ai-trades
// @Param payload body CreateAITradeRequest true "AI trade payload"
// @Success 200 {object} AITradeResponseEnvelope
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/ai-trades [post]
func (h *AITradeHandler) CreateAITrade(c *gin.Context) {
	var req CreateAITradeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TokenAddress == "" || req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	trade := &model.AITrade{
		UserID:         req.UserID,
		TokenAddress:   req.TokenAddress,
		TokenSymbol:    req.TokenSymbol,
		Type:           req.Type,
		AmountIn:       req.AmountIn,
		AmountOut:      req.AmountOut,
		TxHash:         req.TxHash,
		GoldenDogScore: req.GoldenDogScore,
		DecisionReason: req.DecisionReason,
		StrategyUsed:   req.StrategyUsed,
		CurrentValue:   req.CurrentValue,
		ProfitLoss:     req.ProfitLoss,
	}
	if err := h.repo.CreateAITrade(c.Request.Context(), trade); err != nil {
		log.Printf("create ai trade: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": trade})
}

// GetAITrades godoc
// @Summary Get AI trades
// @Description List AI trades
// @Tags ai-trades
// @Param limit query int false "Limit" default(50)
// @Success 200 {object} AITradeListResponseEnvelope
// @Failure 500 {object} map[string]string
// @Router /api/ai-trades [get]
func (h *AITradeHandler) GetAITrades(c *gin.Context) {
	limit := 50
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	trades, err := h.repo.GetAITrades(c.Request.Context(), limit)
	if err != nil {
		log.Printf("get ai trades: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trades})
}

type AITradeStatsResponse struct {
	Count      int64          `json:"count"`
	WinRate    float64        `json:"winRate"`
	AvgPL      float64        `json:"avgPL"`
	TotalPL    float64        `json:"totalPL"`
	ByStrategy []StrategyStat `json:"byStrategy"`
	ByPeriod   []PeriodStat   `json:"byPeriod"`
}

type StrategyStat struct {
	Strategy string  `json:"strategy"`
	Count    int64   `json:"count"`
	WinRate  float64 `json:"winRate"`
	AvgPL    float64 `json:"avgPL"`
	TotalPL  float64 `json:"totalPL"`
}

type PeriodStat struct {
	Period  string  `json:"period"`
	Count   int64   `json:"count"`
	WinRate float64 `json:"winRate"`
	AvgPL   float64 `json:"avgPL"`
	TotalPL float64 `json:"totalPL"`
}

// GetAITradeStats godoc
// @Summary Get AI trade stats
// @Description Aggregate AI trade statistics
// @Tags ai-trades
// @Success 200 {object} map[string]AITradeStatsResponse
// @Failure 500 {object} map[string]string
// @Router /api/ai-trades/stats [get]
func (h *AITradeHandler) GetAITradeStats(c *gin.Context) {
	count, winRate, avgPL, err := h.repo.GetAITradeStats(c.Request.Context())
	if err != nil {
		log.Printf("get ai trade stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	trades, _ := h.repo.GetAllAITrades(c.Request.Context())
	byStrategy := aggregateByStrategy(trades)
	byPeriod := aggregateByPeriod(trades)
	totalPL := 0.0
	for _, t := range trades {
		totalPL += t.ProfitLoss
	}

	c.JSON(http.StatusOK, gin.H{"data": AITradeStatsResponse{
		Count:      count,
		WinRate:    winRate,
		AvgPL:      avgPL,
		TotalPL:    totalPL,
		ByStrategy: byStrategy,
		ByPeriod:   byPeriod,
	}})
}

func aggregateByStrategy(trades []model.AITrade) []StrategyStat {
	type agg struct {
		count int64
		wins  int64
		sumPL float64
	}
	m := map[string]*agg{}
	for _, t := range trades {
		key := t.StrategyUsed
		if key == "" {
			key = "unknown"
		}
		if _, ok := m[key]; !ok {
			m[key] = &agg{}
		}
		m[key].count++
		if t.ProfitLoss > 0 {
			m[key].wins++
		}
		m[key].sumPL += t.ProfitLoss
	}
	out := make([]StrategyStat, 0, len(m))
	for key, a := range m {
		winRate := 0.0
		if a.count > 0 {
			winRate = float64(a.wins) / float64(a.count)
		}
		avg := 0.0
		if a.count > 0 {
			avg = a.sumPL / float64(a.count)
		}
		out = append(out, StrategyStat{
			Strategy: key,
			Count:    a.count,
			WinRate:  winRate,
			AvgPL:    avg,
			TotalPL:  a.sumPL,
		})
	}
	return out
}

func aggregateByPeriod(trades []model.AITrade) []PeriodStat {
	type agg struct {
		count int64
		wins  int64
		sumPL float64
	}
	m := map[string]*agg{}
	for _, t := range trades {
		period := t.Timestamp.Format("2006-01-02")
		if _, ok := m[period]; !ok {
			m[period] = &agg{}
		}
		m[period].count++
		if t.ProfitLoss > 0 {
			m[period].wins++
		}
		m[period].sumPL += t.ProfitLoss
	}
	out := make([]PeriodStat, 0, len(m))
	for period, a := range m {
		winRate := 0.0
		if a.count > 0 {
			winRate = float64(a.wins) / float64(a.count)
		}
		avg := 0.0
		if a.count > 0 {
			avg = a.sumPL / float64(a.count)
		}
		out = append(out, PeriodStat{
			Period:  period,
			Count:   a.count,
			WinRate: winRate,
			AvgPL:   avg,
			TotalPL: a.sumPL,
		})
	}
	return out
}
