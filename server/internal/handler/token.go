package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"easymeme/internal/repository"

	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	repo *repository.Repository
}

func NewTokenHandler(repo *repository.Repository) *TokenHandler {
	return &TokenHandler{repo: repo}
}

func (h *TokenHandler) GetTokens(c *gin.Context) {
	tokens, err := h.repo.GetLatestTokens(c.Request.Context(), 50)
	if err != nil {
		log.Printf("get tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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

type PendingTokenResponse struct {
	Address        string    `json:"address"`
	Name           string    `json:"name"`
	Symbol         string    `json:"symbol"`
	Liquidity      float64   `json:"liquidity"`
	CreatorAddress string    `json:"creatorAddress"`
	CreatedAt      time.Time `json:"createdAt"`
	PairAddress    string    `json:"pairAddress"`
}

func (h *TokenHandler) GetPendingTokens(c *gin.Context) {
	limit := 10
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}

	minLiquidity := 0.0
	if v := c.Query("min_liquidity"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			minLiquidity = parsed
		}
	}

	tokens, err := h.repo.GetPendingTokens(c.Request.Context(), limit, minLiquidity)
	if err != nil {
		log.Printf("get pending tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]PendingTokenResponse, 0, len(tokens))
	for _, token := range tokens {
		resp = append(resp, PendingTokenResponse{
			Address:        token.Address,
			Name:           token.Name,
			Symbol:         token.Symbol,
			Liquidity:      token.InitialLiquidity.InexactFloat64(),
			CreatorAddress: token.CreatorAddress,
			CreatedAt:      token.CreatedAt,
			PairAddress:    token.PairAddress,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *TokenHandler) GetAnalyzedTokens(c *gin.Context) {
	limit := 20
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}

	tokens, err := h.repo.GetAnalyzedTokens(c.Request.Context(), limit)
	if err != nil {
		log.Printf("get analyzed tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tokens})
}

type AnalyzeTokenRiskPayload struct {
	RiskScore      int    `json:"riskScore"`
	RiskLevel      string `json:"riskLevel"`
	IsGoldenDog    bool   `json:"isGoldenDog"`
	Reasoning      string `json:"reasoning"`
	Recommendation string `json:"recommendation"`
	RiskFactors    struct {
		HoneypotRisk      string `json:"honeypotRisk"`
		TaxRisk           string `json:"taxRisk"`
		OwnerRisk         string `json:"ownerRisk"`
		ConcentrationRisk string `json:"concentrationRisk"`
	} `json:"riskFactors"`
}

func (h *TokenHandler) PostTokenAnalysis(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token address is required"})
		return
	}

	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	getString := func(keys ...string) string {
		for _, key := range keys {
			if v, ok := raw[key]; ok {
				if s, ok := v.(string); ok {
					return s
				}
			}
		}
		return ""
	}

	getBool := func(keys ...string) bool {
		for _, key := range keys {
			if v, ok := raw[key]; ok {
				if b, ok := v.(bool); ok {
					return b
				}
			}
		}
		return false
	}

	getInt := func(keys ...string) int {
		for _, key := range keys {
			if v, ok := raw[key]; ok {
				switch typed := v.(type) {
				case float64:
					return int(typed)
				case int:
					return typed
				}
			}
		}
		return 0
	}

	payload := AnalyzeTokenRiskPayload{
		RiskScore:      getInt("riskScore", "risk_score"),
		RiskLevel:      getString("riskLevel", "risk_level"),
		IsGoldenDog:    getBool("isGoldenDog", "is_golden_dog"),
		Reasoning:      getString("reasoning"),
		Recommendation: getString("recommendation"),
	}

	if rf, ok := raw["riskFactors"].(map[string]interface{}); ok {
		if v, ok := rf["honeypotRisk"].(string); ok {
			payload.RiskFactors.HoneypotRisk = v
		}
		if v, ok := rf["taxRisk"].(string); ok {
			payload.RiskFactors.TaxRisk = v
		}
		if v, ok := rf["ownerRisk"].(string); ok {
			payload.RiskFactors.OwnerRisk = v
		}
		if v, ok := rf["concentrationRisk"].(string); ok {
			payload.RiskFactors.ConcentrationRisk = v
		}
	}

	if payload.RiskLevel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "riskLevel is required"})
		return
	}
	if payload.RiskScore < 0 || payload.RiskScore > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "riskScore must be 0-100"})
		return
	}
	validLevels := map[string]bool{"safe": true, "warning": true, "danger": true}
	if !validLevels[strings.ToLower(payload.RiskLevel)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid riskLevel"})
		return
	}

	now := time.Now().UTC()
	riskLevel := strings.ToLower(payload.RiskLevel)
	riskDetails := map[string]interface{}{
		"risk_factors":   payload.RiskFactors,
		"reasoning":      payload.Reasoning,
		"recommendation": payload.Recommendation,
		"is_golden_dog":  payload.IsGoldenDog,
	}
	riskDetailsJSON, _ := json.Marshal(riskDetails)
	analysisJSON, _ := json.Marshal(raw)

	updates := map[string]interface{}{
		"risk_score":      payload.RiskScore,
		"risk_level":      riskLevel,
		"risk_details":    riskDetailsJSON,
		"analysis_result": analysisJSON,
		"analysis_status": "analyzed",
		"is_golden_dog":   payload.IsGoldenDog,
		"analyzed_at":     now,
	}

	if payload.RiskFactors.HoneypotRisk != "" {
		updates["is_honeypot"] = strings.EqualFold(payload.RiskFactors.HoneypotRisk, "HIGH")
	}

	if err := h.repo.UpdateTokenAnalysis(c.Request.Context(), address, updates); err != nil {
		log.Printf("update token analysis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
