package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"easymeme/internal/model"
	"easymeme/internal/repository"

	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	repo *repository.Repository
}

func NewTokenHandler(repo *repository.Repository) *TokenHandler {
	return &TokenHandler{repo: repo}
}

func toTokenDTO(token model.Token) TokenResponseDTO {
	return TokenResponseDTO{
		ID:               token.ID,
		Address:          token.Address,
		Name:             token.Name,
		Symbol:           token.Symbol,
		Decimals:         token.Decimals,
		PairAddress:      token.PairAddress,
		Dex:              token.Dex,
		InitialLiquidity: token.InitialLiquidity.String(),
		AnalysisStatus:   token.AnalysisStatus,
		RiskScore:        token.RiskScore,
		RiskLevel:        token.RiskLevel,
		IsGoldenDog:      token.IsGoldenDog,
		GoldenDogScore:   token.GoldenDogScore,
		IsHoneypot:       token.IsHoneypot,
		BuyTax:           token.BuyTax,
		SellTax:          token.SellTax,
		CreatorAddress:   token.CreatorAddress,
		CreatedAt:        token.CreatedAt,
		UpdatedAt:        token.UpdatedAt,
		AnalyzedAt:       token.AnalyzedAt,
	}
}

func toTokenDTOList(tokens []model.Token) []TokenResponseDTO {
	resp := make([]TokenResponseDTO, 0, len(tokens))
	for _, token := range tokens {
		resp = append(resp, toTokenDTO(token))
	}
	return resp
}

type TokenResponseDTO struct {
	ID               string     `json:"id"`
	Address          string     `json:"address"`
	Name             string     `json:"name"`
	Symbol           string     `json:"symbol"`
	Decimals         int        `json:"decimals"`
	PairAddress      string     `json:"pair_address"`
	Dex              string     `json:"dex"`
	InitialLiquidity string     `json:"initial_liquidity"`
	AnalysisStatus   string     `json:"analysis_status"`
	RiskScore        int        `json:"risk_score"`
	RiskLevel        string     `json:"risk_level"`
	IsGoldenDog      bool       `json:"is_golden_dog"`
	GoldenDogScore   int        `json:"golden_dog_score"`
	IsHoneypot       bool       `json:"is_honeypot"`
	BuyTax           float64    `json:"buy_tax"`
	SellTax          float64    `json:"sell_tax"`
	CreatorAddress   string     `json:"creator_address"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	AnalyzedAt       *time.Time `json:"analyzed_at"`
}

type TokenListResponseEnvelope struct {
	Data []TokenResponseDTO `json:"data"`
}

// GetTokens godoc
// @Summary Get latest tokens
// @Description List latest tokens
// @Tags tokens
// @Success 200 {object} TokenListResponseEnvelope
// @Failure 500 {object} map[string]string
// @Router /api/tokens [get]
func (h *TokenHandler) GetTokens(c *gin.Context) {
	tokens, err := h.repo.GetLatestTokens(c.Request.Context(), 50)
	if err != nil {
		log.Printf("get tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toTokenDTOList(tokens)})
}

type TokenResponseEnvelope struct {
	Data TokenResponseDTO `json:"data"`
}

// GetToken godoc
// @Summary Get token
// @Description Get token by address
// @Tags tokens
// @Param address path string true "Token address"
// @Success 200 {object} TokenResponseEnvelope
// @Failure 404 {object} map[string]string
// @Router /api/tokens/{address} [get]
func (h *TokenHandler) GetToken(c *gin.Context) {
	address := c.Param("address")
	token, err := h.repo.GetTokenByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toTokenDTO(*token)})
}

type TokenDetailResponse struct {
	ID                 string                 `json:"id"`
	Address            string                 `json:"address"`
	Name               string                 `json:"name"`
	Symbol             string                 `json:"symbol"`
	PairAddress        string                 `json:"pairAddress"`
	Dex                string                 `json:"dex"`
	Liquidity          float64                `json:"liquidity"`
	CreatorAddress     string                 `json:"creatorAddress"`
	CreatedAt          time.Time              `json:"createdAt"`
	AnalyzedAt         *time.Time             `json:"analyzedAt"`
	RiskScore          int                    `json:"riskScore"`
	RiskLevel          string                 `json:"riskLevel"`
	IsGoldenDog        bool                   `json:"isGoldenDog"`
	GoldenDogScore     int                    `json:"goldenDogScore"`
	EffectiveScore     int                    `json:"effectiveScore"`
	TimeDecayFactor    float64                `json:"timeDecayFactor"`
	Phase              string                 `json:"phase"`
	RiskDetails        map[string]interface{} `json:"riskDetails,omitempty"`
	AnalysisResult     map[string]interface{} `json:"analysisResult,omitempty"`
	GoPlus             any                    `json:"goplus"`
	DEXScreener        any                    `json:"dexscreener"`
	HolderDistribution any                    `json:"holderDistribution"`
	CreatorHistory     any                    `json:"creatorHistory"`
	MarketAlerts       any                    `json:"marketAlerts"`
}

type TokenDetailResponseEnvelope struct {
	Data TokenDetailResponse `json:"data"`
}

// GetTokenDetail godoc
// @Summary Get token detail
// @Description Get token detail including scores and time decay metadata
// @Tags tokens
// @Param address path string true "Token address"
// @Success 200 {object} TokenDetailResponseEnvelope
// @Failure 404 {object} map[string]string
// @Router /api/tokens/{address}/detail [get]
func (h *TokenHandler) GetTokenDetail(c *gin.Context) {
	address := c.Param("address")
	token, err := h.repo.GetTokenByAddress(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}

	var riskDetails map[string]interface{}
	if len(token.RiskDetails) > 0 {
		_ = json.Unmarshal(token.RiskDetails, &riskDetails)
	}
	var analysisResult map[string]interface{}
	if len(token.AnalysisResult) > 0 {
		_ = json.Unmarshal(token.AnalysisResult, &analysisResult)
	}
	goplusData := map[string]interface{}{}
	if len(token.RiskDetails) > 0 {
		if normalized, ok := riskDetails["normalized"].(map[string]interface{}); ok {
			goplusData = normalized
			if raw, exists := riskDetails["raw"]; exists {
				goplusData["raw"] = raw
			}
		} else {
			goplusData = riskDetails
		}
	}
	marketData := map[string]interface{}{}
	if len(token.MarketData) > 0 {
		_ = json.Unmarshal(token.MarketData, &marketData)
	}
	holderData := map[string]interface{}{}
	if len(token.HolderData) > 0 {
		_ = json.Unmarshal(token.HolderData, &holderData)
	}
	creatorData := map[string]interface{}{}
	if len(token.CreatorHistory) > 0 {
		_ = json.Unmarshal(token.CreatorHistory, &creatorData)
	}
	alertData := []map[string]interface{}{}
	if len(token.MarketAlerts) > 0 {
		_ = json.Unmarshal(token.MarketAlerts, &alertData)
	}

	resp := TokenDetailResponse{
		ID:                 token.ID,
		Address:            token.Address,
		Name:               token.Name,
		Symbol:             token.Symbol,
		PairAddress:        token.PairAddress,
		Dex:                token.Dex,
		Liquidity:          token.InitialLiquidity.InexactFloat64(),
		CreatorAddress:     token.CreatorAddress,
		CreatedAt:          token.CreatedAt,
		AnalyzedAt:         token.AnalyzedAt,
		RiskScore:          token.RiskScore,
		RiskLevel:          token.RiskLevel,
		IsGoldenDog:        token.IsGoldenDog,
		GoldenDogScore:     token.GoldenDogScore,
		EffectiveScore:     token.EffectiveScore(),
		TimeDecayFactor:    token.TimeDecayFactor(),
		Phase:              token.GoldenDogPhase(),
		RiskDetails:        riskDetails,
		AnalysisResult:     analysisResult,
		GoPlus:             goplusData,
		DEXScreener:        marketData,
		HolderDistribution: holderData,
		CreatorHistory:     creatorData,
		MarketAlerts:       alertData,
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

type PendingTokenResponse struct {
	Address            string    `json:"address"`
	Name               string    `json:"name"`
	Symbol             string    `json:"symbol"`
	Liquidity          float64   `json:"liquidity"`
	CreatorAddress     string    `json:"creatorAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	PairAddress        string    `json:"pairAddress"`
	GoPlus             any       `json:"goplus"`
	DEXScreener        any       `json:"dexscreener"`
	HolderDistribution any       `json:"holderDistribution"`
	CreatorHistory     any       `json:"creatorHistory"`
	MarketAlerts       any       `json:"marketAlerts"`
	SocialSignals      any       `json:"socialSignals"`
	SmartMoneySignals  any       `json:"smartMoneySignals"`
}

type PendingTokenListResponseEnvelope struct {
	Data []PendingTokenResponse `json:"data"`
}

// GetPendingTokens godoc
// @Summary Get pending tokens
// @Description List tokens pending analysis
// @Tags tokens
// @Param limit query int false "Limit" default(10)
// @Param min_liquidity query number false "Min liquidity"
// @Success 200 {object} PendingTokenListResponseEnvelope
// @Failure 500 {object} map[string]string
// @Router /api/tokens/pending [get]
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
		goplusData := map[string]interface{}{}
		if len(token.RiskDetails) > 0 {
			var details map[string]interface{}
			_ = json.Unmarshal(token.RiskDetails, &details)
			if normalized, ok := details["normalized"].(map[string]interface{}); ok {
				goplusData = normalized
				if raw, exists := details["raw"]; exists {
					goplusData["raw"] = raw
				}
			} else {
				goplusData = details
			}
		}
		marketData := map[string]interface{}{}
		if len(token.MarketData) > 0 {
			_ = json.Unmarshal(token.MarketData, &marketData)
		}
		holderData := map[string]interface{}{}
		if len(token.HolderData) > 0 {
			_ = json.Unmarshal(token.HolderData, &holderData)
		}
		creatorData := map[string]interface{}{}
		if len(token.CreatorHistory) > 0 {
			_ = json.Unmarshal(token.CreatorHistory, &creatorData)
		}
		alertData := []map[string]interface{}{}
		if len(token.MarketAlerts) > 0 {
			_ = json.Unmarshal(token.MarketAlerts, &alertData)
		}
		socialSignals := map[string]interface{}{}
		if len(token.SocialSignals) > 0 {
			_ = json.Unmarshal(token.SocialSignals, &socialSignals)
		}
		smartMoneySignals := map[string]interface{}{}
		if len(token.SmartMoneySignals) > 0 {
			_ = json.Unmarshal(token.SmartMoneySignals, &smartMoneySignals)
		}
		resp = append(resp, PendingTokenResponse{
			Address:            token.Address,
			Name:               token.Name,
			Symbol:             token.Symbol,
			Liquidity:          token.InitialLiquidity.InexactFloat64(),
			CreatorAddress:     token.CreatorAddress,
			CreatedAt:          token.CreatedAt,
			PairAddress:        token.PairAddress,
			GoPlus:             goplusData,
			DEXScreener:        marketData,
			HolderDistribution: holderData,
			CreatorHistory:     creatorData,
			MarketAlerts:       alertData,
			SocialSignals:      socialSignals,
			SmartMoneySignals:  smartMoneySignals,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

type AnalyzedTokenListResponseEnvelope struct {
	Data []TokenResponseDTO `json:"data"`
}

// GetAnalyzedTokens godoc
// @Summary Get analyzed tokens
// @Description List analyzed tokens
// @Tags tokens
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} AnalyzedTokenListResponseEnvelope
// @Failure 500 {object} map[string]string
// @Router /api/tokens/analyzed [get]
func (h *TokenHandler) GetAnalyzedTokens(c *gin.Context) {
	limit := 20
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}

	if daysQuery := c.Query("days"); daysQuery != "" {
		days := 7
		if parsed, err := strconv.Atoi(daysQuery); err == nil && parsed > 0 {
			days = parsed
		}
		page := 1
		if v := c.Query("page"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				page = parsed
			}
		}
		pageSize := limit
		if v := c.Query("pageSize"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				pageSize = parsed
			}
		}
		offset := (page - 1) * pageSize
		since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		tokens, total, err := h.repo.GetAnalyzedTokensSince(c.Request.Context(), since, pageSize, offset)
		if err != nil {
			log.Printf("get analyzed tokens by days: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":     toTokenDTOList(tokens),
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
			"since":    since,
		})
		return
	}

	tokens, err := h.repo.GetAnalyzedTokens(c.Request.Context(), limit)
	if err != nil {
		log.Printf("get analyzed tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toTokenDTOList(tokens)})
}

type GoldenDogResponse struct {
	Address         string     `json:"address"`
	Name            string     `json:"name"`
	Symbol          string     `json:"symbol"`
	PairAddress     string     `json:"pairAddress"`
	Liquidity       float64    `json:"liquidity"`
	RiskScore       int        `json:"riskScore"`
	RiskLevel       string     `json:"riskLevel"`
	IsGoldenDog     bool       `json:"isGoldenDog"`
	GoldenDogScore  int        `json:"goldenDogScore"`
	EffectiveScore  int        `json:"effectiveScore"`
	TimeDecayFactor float64    `json:"timeDecayFactor"`
	Phase           string     `json:"phase"`
	CreatedAt       time.Time  `json:"createdAt"`
	AnalyzedAt      *time.Time `json:"analyzedAt"`
}

type GoldenDogListResponseEnvelope struct {
	Data []GoldenDogResponse `json:"data"`
}

// GetGoldenDogs godoc
// @Summary Get golden dog tokens
// @Description List golden dog tokens sorted by effective score and excluding expired
// @Tags tokens
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} GoldenDogListResponseEnvelope
// @Failure 500 {object} map[string]string
// @Router /api/tokens/golden-dogs [get]
func (h *TokenHandler) GetGoldenDogs(c *gin.Context) {
	limit := 20
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}

	fetchLimit := limit * 5
	if fetchLimit < 20 {
		fetchLimit = 20
	}
	if fetchLimit > 200 {
		fetchLimit = 200
	}

	tokens, err := h.repo.GetGoldenDogTokens(c.Request.Context(), fetchLimit)
	if err != nil {
		log.Printf("get golden dog tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]GoldenDogResponse, 0, len(tokens))
	for _, token := range tokens {
		if token.GoldenDogPhase() == "EXPIRED" {
			continue
		}
		resp = append(resp, GoldenDogResponse{
			Address:         token.Address,
			Name:            token.Name,
			Symbol:          token.Symbol,
			PairAddress:     token.PairAddress,
			Liquidity:       token.InitialLiquidity.InexactFloat64(),
			RiskScore:       token.RiskScore,
			RiskLevel:       token.RiskLevel,
			IsGoldenDog:     token.IsGoldenDog,
			GoldenDogScore:  token.GoldenDogScore,
			EffectiveScore:  token.EffectiveScore(),
			TimeDecayFactor: token.TimeDecayFactor(),
			Phase:           token.GoldenDogPhase(),
			CreatedAt:       token.CreatedAt,
			AnalyzedAt:      token.AnalyzedAt,
		})
	}

	sort.Slice(resp, func(i, j int) bool {
		if resp[i].EffectiveScore == resp[j].EffectiveScore {
			left := resp[i].AnalyzedAt
			right := resp[j].AnalyzedAt
			if left == nil && right == nil {
				return false
			}
			if left == nil {
				return false
			}
			if right == nil {
				return true
			}
			return left.After(*right)
		}
		return resp[i].EffectiveScore > resp[j].EffectiveScore
	})

	if limit > 0 && len(resp) > limit {
		resp = resp[:limit]
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

type GoldenDogScoreBucket struct {
	Range string `json:"range"`
	Count int64  `json:"count"`
}

// GetGoldenDogScoreDistribution godoc
// @Summary Get goldenDogScore distribution
// @Description Get analyzed token goldenDogScore distribution for recent days
// @Tags tokens
// @Param days query int false "Recent days" default(7)
// @Param bucket query int false "Bucket size" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/tokens/stats/golden-dog-score-distribution [get]
func (h *TokenHandler) GetGoldenDogScoreDistribution(c *gin.Context) {
	days := 7
	if v := c.Query("days"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			days = parsed
		}
	}
	bucket := 10
	if v := c.Query("bucket"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			bucket = parsed
		}
	}
	since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	raw, total, err := h.repo.GetGoldenDogScoreDistributionSince(c.Request.Context(), since, bucket)
	if err != nil {
		log.Printf("get goldenDogScore distribution: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	maxBucket := 100 / bucket
	buckets := make([]GoldenDogScoreBucket, 0, maxBucket+1)
	for i := 0; i <= maxBucket; i++ {
		start := i * bucket
		end := start + bucket - 1
		if end > 100 {
			end = 100
		}
		buckets = append(buckets, GoldenDogScoreBucket{
			Range: strconv.Itoa(start) + "-" + strconv.Itoa(end),
			Count: raw[i],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"since":         since,
		"until":         time.Now().UTC(),
		"bucketSize":    bucket,
		"totalAnalyzed": total,
		"distribution":  buckets,
	})
}

type TokenPricePoint struct {
	TS           time.Time `json:"ts"`
	PriceUSD     float64   `json:"price"`
	LiquidityUSD float64   `json:"liquidityUsd,omitempty"`
	Volume5mUSD  float64   `json:"volume5mUsd,omitempty"`
}

// GetTokenPriceSeries godoc
// @Summary Get token price series
// @Description Get analyzed token subsequent price series
// @Tags tokens
// @Param address path string true "Token address"
// @Param from query string false "RFC3339 start time"
// @Param to query string false "RFC3339 end time"
// @Param limit query int false "Max points" default(2000)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/tokens/{address}/price-series [get]
func (h *TokenHandler) GetTokenPriceSeries(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token address is required"})
		return
	}
	to := time.Now().UTC()
	from := to.Add(-24 * time.Hour)
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	limit := 2000
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	rows, err := h.repo.GetTokenPriceSeries(c.Request.Context(), address, from, to, limit)
	if err != nil {
		log.Printf("get token price series: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	series := make([]TokenPricePoint, 0, len(rows))
	for _, r := range rows {
		series = append(series, TokenPricePoint{TS: r.TS, PriceUSD: r.PriceUSD, LiquidityUSD: r.LiquidityUSD, Volume5mUSD: r.Volume5mUSD})
	}
	c.JSON(http.StatusOK, gin.H{
		"tokenAddress": address,
		"from":         from,
		"to":           to,
		"count":        len(series),
		"series":       series,
	})
}

type UpsertTokenPriceSnapshotPayload struct {
	TokenAddress string  `json:"tokenAddress"`
	TS           string  `json:"ts"`
	PriceUSD     float64 `json:"priceUsd"`
	LiquidityUSD float64 `json:"liquidityUsd"`
	Volume5mUSD  float64 `json:"volume5mUsd"`
}

// UpsertTokenPriceSnapshot godoc
// @Summary Upsert token price snapshot
// @Description Upsert a token price snapshot (for tx_agent data feed)
// @Tags tokens
// @Param payload body UpsertTokenPriceSnapshotPayload true "snapshot"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tokens/price-snapshots [post]
func (h *TokenHandler) UpsertTokenPriceSnapshot(c *gin.Context) {
	var payload UpsertTokenPriceSnapshotPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if payload.TokenAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tokenAddress is required"})
		return
	}
	if payload.PriceUSD <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "priceUsd must be > 0"})
		return
	}
	ts := time.Now().UTC()
	if payload.TS != "" {
		parsed, err := time.Parse(time.RFC3339, payload.TS)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ts must be RFC3339"})
			return
		}
		ts = parsed
	}
	row := &model.TokenPriceSnapshot{
		TokenAddress: payload.TokenAddress,
		TS:           ts,
		PriceUSD:     payload.PriceUSD,
		LiquidityUSD: payload.LiquidityUSD,
		Volume5mUSD:  payload.Volume5mUSD,
	}
	if err := h.repo.UpsertTokenPriceSnapshot(c.Request.Context(), row); err != nil {
		log.Printf("upsert token price snapshot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type AnalyzeTokenRiskPayload struct {
	RiskScore      int    `json:"riskScore"`
	RiskLevel      string `json:"riskLevel"`
	IsGoldenDog    bool   `json:"isGoldenDog"`
	GoldenDogScore int    `json:"goldenDogScore"`
	Reasoning      string `json:"reasoning"`
	Recommendation string `json:"recommendation"`
	RiskFactors    struct {
		HoneypotRisk      string `json:"honeypotRisk"`
		TaxRisk           string `json:"taxRisk"`
		OwnerRisk         string `json:"ownerRisk"`
		ConcentrationRisk string `json:"concentrationRisk"`
	} `json:"riskFactors"`
}

type AnalysisStatusResponseEnvelope struct {
	Status string `json:"status"`
}

// PostTokenAnalysis godoc
// @Summary Submit token analysis
// @Description Submit AI analysis for a token
// @Tags tokens
// @Param address path string true "Token address"
// @Param payload body AnalyzeTokenRiskPayload true "Analysis payload"
// @Success 200 {object} AnalysisStatusResponseEnvelope
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/tokens/{address}/analysis [post]
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

	getIntWithPresence := func(keys ...string) (int, bool) {
		for _, key := range keys {
			if v, ok := raw[key]; ok {
				switch typed := v.(type) {
				case float64:
					return int(typed), true
				case int:
					return typed, true
				}
				return 0, true
			}
		}
		return 0, false
	}

	payload := AnalyzeTokenRiskPayload{
		RiskScore:      getInt("riskScore", "risk_score"),
		RiskLevel:      getString("riskLevel", "risk_level"),
		IsGoldenDog:    getBool("isGoldenDog", "is_golden_dog"),
		GoldenDogScore: getInt("goldenDogScore", "golden_dog_score"),
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
	if score, ok := getIntWithPresence("goldenDogScore", "golden_dog_score"); ok {
		if score < 0 || score > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "goldenDogScore must be 0-100"})
			return
		}
		payload.GoldenDogScore = score
	}
	validLevels := map[string]bool{"safe": true, "warning": true, "danger": true}
	if !validLevels[strings.ToLower(payload.RiskLevel)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid riskLevel"})
		return
	}

	now := time.Now().UTC()
	riskLevel := strings.ToLower(payload.RiskLevel)
	analysisJSON, _ := json.Marshal(raw)

	updates := map[string]interface{}{
		"risk_score":       payload.RiskScore,
		"risk_level":       riskLevel,
		"analysis_result":  analysisJSON,
		"analysis_status":  "analyzed",
		"is_golden_dog":    payload.IsGoldenDog,
		"golden_dog_score": payload.GoldenDogScore,
		"analyzed_at":      now,
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
