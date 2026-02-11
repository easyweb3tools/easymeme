package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	basesdk "easyweb3/go-sdk"

	"easyweb3/apps/easymeme/internal/model"
	"easyweb3/apps/easymeme/internal/repository"
	"easyweb3/apps/easymeme/pkg/ethereum"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Scanner struct {
	client *ethereum.Client
	repo   *repository.Repository
	hub    Broadcaster
	base   *basesdk.Client
	stats  *enrichmentStats
}

type Broadcaster interface {
	Broadcast(payload interface{})
}

type EnrichmentStatsSnapshot struct {
	EnrichSuccess       int64     `json:"enrich_success"`
	EnrichFailure       int64     `json:"enrich_failure"`
	RefreshSuccess      int64     `json:"refresh_success"`
	RefreshFailure      int64     `json:"refresh_failure"`
	LastEnrichSuccessAt time.Time `json:"last_enrich_success_at"`
	LastEnrichFailureAt time.Time `json:"last_enrich_failure_at"`
	LastRefreshAt       time.Time `json:"last_refresh_at"`
	LastErrors          []string  `json:"last_errors"`
}

type enrichmentStats struct {
	mu                  sync.Mutex
	enrichSuccess       int64
	enrichFailure       int64
	refreshSuccess      int64
	refreshFailure      int64
	lastEnrichSuccessAt time.Time
	lastEnrichFailureAt time.Time
	lastRefreshAt       time.Time
	lastErrors          []string
}

func newEnrichmentStats() *enrichmentStats {
	return &enrichmentStats{lastErrors: make([]string, 0, 20)}
}

func (s *enrichmentStats) recordEnrichSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enrichSuccess++
	s.lastEnrichSuccessAt = time.Now().UTC()
}

func (s *enrichmentStats) recordEnrichFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enrichFailure++
	s.lastEnrichFailureAt = time.Now().UTC()
	s.pushError(err)
}

func (s *enrichmentStats) recordRefreshSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshSuccess++
	s.lastRefreshAt = time.Now().UTC()
}

func (s *enrichmentStats) recordRefreshFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshFailure++
	s.lastRefreshAt = time.Now().UTC()
	s.pushError(err)
}

func (s *enrichmentStats) pushError(err error) {
	if err == nil {
		return
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return
	}
	s.lastErrors = append(s.lastErrors, msg)
	if len(s.lastErrors) > 20 {
		s.lastErrors = s.lastErrors[len(s.lastErrors)-20:]
	}
}

func (s *enrichmentStats) snapshot() EnrichmentStatsSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyErrs := make([]string, len(s.lastErrors))
	copy(copyErrs, s.lastErrors)
	return EnrichmentStatsSnapshot{
		EnrichSuccess:       s.enrichSuccess,
		EnrichFailure:       s.enrichFailure,
		RefreshSuccess:      s.refreshSuccess,
		RefreshFailure:      s.refreshFailure,
		LastEnrichSuccessAt: s.lastEnrichSuccessAt,
		LastEnrichFailureAt: s.lastEnrichFailureAt,
		LastRefreshAt:       s.lastRefreshAt,
		LastErrors:          copyErrs,
	}
}

func NewScanner(client *ethereum.Client, repo *repository.Repository, hub Broadcaster, baseClient *basesdk.Client) *Scanner {
	return &Scanner{
		client: client,
		repo:   repo,
		hub:    hub,
		base:   baseClient,
		stats:  newEnrichmentStats(),
	}
}

func (s *Scanner) Start(ctx context.Context) error {
	logs, sub, err := s.client.SubscribePairCreated(ctx)
	if err != nil {
		log.Printf("[Scanner] Subscription unavailable: %v", err)
		go s.pollPairCreated(ctx)
	} else {
		log.Println("[Scanner] Started listening for PairCreated events...")
		go func() {
			for {
				select {
				case err := <-sub.Err():
					log.Printf("[Scanner] Subscription error: %v", err)
					return
				case vLog := <-logs:
					go s.handlePairCreated(ctx, vLog)
				case <-ctx.Done():
					log.Println("[Scanner] Stopping...")
					return
				}
			}
		}()
	}

	go s.recoverEnrichmentLoop(ctx)
	go s.refreshMarketLoop(ctx)
	go s.logStatsLoop(ctx)
	return nil
}

func (s *Scanner) HealthStatus() map[string]interface{} {
	stats := s.stats.snapshot()
	return map[string]interface{}{
		"enrichment": stats,
		"dependencies": map[string]interface{}{
			"base_service": "configured",
		},
	}
}

func (s *Scanner) pollPairCreated(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var lastBlock uint64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest, err := s.client.LatestBlockNumber(ctx)
			if err != nil {
				log.Printf("[Scanner] BlockNumber error: %v", err)
				continue
			}

			var from uint64
			if lastBlock == 0 {
				if latest > 5000 {
					from = latest - 5000
				} else {
					from = 0
				}
			} else {
				from = lastBlock + 1
			}

			if from > latest {
				continue
			}

			logs, err := s.client.GetPairCreatedLogs(ctx, from, latest)
			if err != nil {
				log.Printf("[Scanner] Poll logs error: %v", err)
				continue
			}

			for _, vLog := range logs {
				s.handlePairCreated(ctx, vLog)
			}

			lastBlock = latest
		}
	}
}

func (s *Scanner) handlePairCreated(ctx context.Context, vLog types.Log) {
	token0 := common.HexToAddress(vLog.Topics[1].Hex())
	token1 := common.HexToAddress(vLog.Topics[2].Hex())
	if len(vLog.Data) < 32 {
		return
	}
	pairAddr := common.BytesToAddress(vLog.Data[:32])

	wbnb := common.HexToAddress(ethereum.WBNB)
	var targetToken common.Address
	if strings.EqualFold(token0.Hex(), wbnb.Hex()) {
		targetToken = token1
	} else if strings.EqualFold(token1.Hex(), wbnb.Hex()) {
		targetToken = token0
	} else {
		return
	}

	log.Printf("[Scanner] New pair: %s, Token: %s", pairAddr.Hex(), targetToken.Hex())

	if s.repo.TokenExists(ctx, targetToken.Hex()) {
		return
	}

	name, symbol, decimals, _ := s.client.GetTokenInfo(ctx, targetToken)

	reserve0, reserve1, _ := s.client.GetPairReserves(ctx, pairAddr)
	var liquidity *big.Int
	if strings.EqualFold(token0.Hex(), wbnb.Hex()) {
		liquidity = reserve0
	} else {
		liquidity = reserve1
	}
	if liquidity == nil {
		liquidity = big.NewInt(0)
	}

	token := &model.Token{
		Address:          targetToken.Hex(),
		Name:             name,
		Symbol:           symbol,
		Decimals:         int(decimals),
		PairAddress:      pairAddr.Hex(),
		Dex:              "pancakeswap",
		InitialLiquidity: decimal.NewFromBigInt(liquidity, -18),
		RiskScore:        0,
		RiskLevel:        "pending",
		AnalysisStatus:   "pending",
		IsGoldenDog:      false,
		IsHoneypot:       false,
		BuyTax:           0,
		SellTax:          0,
	}
	detailsJSON, _ := json.Marshal(map[string]interface{}{"status": "pending"})
	token.RiskDetails = detailsJSON

	if err := s.repo.CreateToken(ctx, token); err != nil {
		log.Printf("[Scanner] Save token error: %v", err)
		return
	}

	go s.enrichTokenWithRetry(ctx, token.Address, token.PairAddress, 3, "new_pair")

	s.hub.Broadcast(map[string]interface{}{
		"type":  "new_token",
		"token": token,
	})

	log.Printf("[Scanner] Token saved: %s (%s), Risk: %d", symbol, targetToken.Hex(), token.RiskScore)
}

func (s *Scanner) enrichTokenWithRetry(ctx context.Context, tokenAddress, pairAddress string, maxAttempts int, reason string) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := s.markEnriching(ctx, tokenAddress, attempt); err != nil {
			log.Printf("[Scanner] mark enriching failed for %s: %v", tokenAddress, err)
		}

		err := s.enrichTokenOnce(ctx, tokenAddress, pairAddress)
		if err == nil {
			s.stats.recordEnrichSuccess()
			return
		}

		lastErr = err
		s.stats.recordEnrichFailure(err)
		log.Printf("[Scanner] enrich attempt failed token=%s attempt=%d reason=%s err=%v", tokenAddress, attempt, reason, err)

		if attempt < maxAttempts {
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return
			}
		}
	}

	if err := s.repo.UpdateTokenAnalysis(ctx, tokenAddress, map[string]interface{}{
		"analysis_status": "enrich_failed",
		"enrich_error":    trimErr(lastErr),
	}); err != nil {
		log.Printf("[Scanner] set enrich_failed failed for %s: %v", tokenAddress, err)
	}
}

func (s *Scanner) markEnriching(ctx context.Context, tokenAddress string, attempt int) error {
	return s.repo.UpdateTokenAnalysis(ctx, tokenAddress, map[string]interface{}{
		"analysis_status": "enriching",
		"enrich_error":    "",
		"enrich_attempts": gorm.Expr("COALESCE(enrich_attempts, 0) + 1"),
	})
}

func (s *Scanner) enrichTokenOnce(ctx context.Context, tokenAddress, pairAddress string) error {
	goplusData, err := s.base.GetTokenSecurity(ctx, "bsc", tokenAddress)
	if err != nil {
		return err
	}

	marketRaw := map[string]interface{}{}
	if pairAddress != "" {
		pairData, derr := s.base.GetPairData(ctx, "bsc", pairAddress)
		if derr != nil {
			log.Printf("[Scanner] DEXScreener enrich warning for %s/%s: %v", tokenAddress, pairAddress, derr)
		} else {
			marketRaw = pairData
		}
	}

	normalizedGoPlus := normalizeGoPlus(goplusData)
	normalizedDex := normalizeDEXScreener(marketRaw)

	holderData := map[string]interface{}{}
	if holders, herr := s.base.GetHolderDistribution(ctx, "bsc", tokenAddress); herr == nil {
		holderData = map[string]interface{}{
			"topHolders": holders.TopHolders,
			"top10Share": holders.Top10Share,
			"total":      holders.Total,
			"source":     holders.Source,
		}
		normalizedGoPlus["top10_holder_share"] = holders.Top10Share
	} else {
		log.Printf("[Scanner] holder distribution warning for %s: %v", tokenAddress, herr)
	}

	creatorHistory := map[string]interface{}{}
	if history, cerr := s.base.GetCreatorHistory(ctx, "bsc", tokenAddress); cerr == nil {
		creatorHistory = map[string]interface{}{
			"creatorAddress":   history.CreatorAddress,
			"contractAddress":  history.ContractAddress,
			"creationTxHash":   history.CreationTxHash,
			"createdContracts": history.CreatedContracts,
			"recentTxs":        history.RecentTxs,
			"source":           history.Source,
		}
		if normalizedGoPlus["creator_address"] == "" {
			normalizedGoPlus["creator_address"] = history.CreatorAddress
		}
	} else {
		log.Printf("[Scanner] creator history warning for %s: %v", tokenAddress, cerr)
	}

	riskDetailsJSON, err := json.Marshal(map[string]interface{}{
		"raw":        goplusData.Raw,
		"normalized": normalizedGoPlus,
	})
	if err != nil {
		return err
	}
	marketDataJSON, err := json.Marshal(normalizedDex)
	if err != nil {
		return err
	}
	holderDataJSON, _ := json.Marshal(holderData)
	creatorHistoryJSON, _ := json.Marshal(creatorHistory)

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"is_honeypot":            parseBinaryFlag(goplusData.IsHoneypot),
		"buy_tax":                parsePercentNumber(goplusData.BuyTax),
		"sell_tax":               parsePercentNumber(goplusData.SellTax),
		"creator_address":        asString(normalizedGoPlus["creator_address"]),
		"risk_details":           riskDetailsJSON,
		"market_data":            marketDataJSON,
		"holder_data":            holderDataJSON,
		"creator_history":        creatorHistoryJSON,
		"analysis_status":        "enriched",
		"enrich_error":           "",
		"enriched_at":            now,
		"last_market_refresh_at": now,
	}

	if err := s.repo.UpdateTokenAnalysis(ctx, tokenAddress, updates); err != nil {
		return err
	}

	if err := s.storeMarketSnapshotAndAlert(ctx, tokenAddress, pairAddress, normalizedDex); err != nil {
		log.Printf("[Scanner] snapshot/alert warning for %s: %v", tokenAddress, err)
	}

	log.Printf("[Scanner] Token enriched: %s", tokenAddress)
	return nil
}

func (s *Scanner) recoverEnrichmentLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.recoverEnrichmentBatch(ctx)
		}
	}
}

func (s *Scanner) recoverEnrichmentBatch(ctx context.Context) {
	recoverStatuses := []string{"pending", "enrich_failed"}
	for _, status := range recoverStatuses {
		tokens, err := s.repo.GetTokensByStatus(ctx, status, 20)
		if err != nil {
			log.Printf("[Scanner] recover list status=%s err=%v", status, err)
			continue
		}
		for _, token := range tokens {
			s.enrichTokenWithRetry(ctx, token.Address, token.PairAddress, 3, "recovery")
			select {
			case <-time.After(500 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}
	}

	staleCutoff := time.Now().UTC().Add(-10 * time.Minute)
	stale, err := s.repo.GetStaleEnrichingTokens(ctx, staleCutoff, 20)
	if err != nil {
		log.Printf("[Scanner] stale enriching list err=%v", err)
		return
	}
	for _, token := range stale {
		s.enrichTokenWithRetry(ctx, token.Address, token.PairAddress, 2, "stale_enriching")
	}
}

func (s *Scanner) refreshMarketLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.refreshMarketBatch(ctx)
		}
	}
}

func (s *Scanner) refreshMarketBatch(ctx context.Context) {
	freshBefore := time.Now().UTC().Add(-5 * time.Minute)
	since := time.Now().UTC().Add(-6 * time.Hour)
	tokens, err := s.repo.GetTokensForMarketRefresh(ctx, freshBefore, since, 120)
	if err != nil {
		log.Printf("[Scanner] market refresh query failed: %v", err)
		s.stats.recordRefreshFailure(err)
		return
	}

	for _, token := range tokens {
		if err := s.refreshTokenMarketData(ctx, token); err != nil {
			s.stats.recordRefreshFailure(err)
			log.Printf("[Scanner] market refresh failed token=%s err=%v", token.Address, err)
		} else {
			s.stats.recordRefreshSuccess()
		}
		select {
		case <-time.After(350 * time.Millisecond):
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scanner) refreshTokenMarketData(ctx context.Context, token model.Token) error {
	if token.PairAddress == "" {
		return errors.New("missing pair address")
	}
	pairData, err := s.base.GetPairData(ctx, "bsc", token.PairAddress)
	if err != nil {
		return err
	}
	normalized := normalizeDEXScreener(pairData)
	marketDataJSON, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := s.repo.UpdateTokenAnalysis(ctx, token.Address, map[string]interface{}{
		"market_data":            marketDataJSON,
		"last_market_refresh_at": now,
	}); err != nil {
		return err
	}

	if err := s.storeMarketSnapshotAndAlert(ctx, token.Address, token.PairAddress, normalized); err != nil {
		return err
	}
	return nil
}

func (s *Scanner) storeMarketSnapshotAndAlert(ctx context.Context, tokenAddress, pairAddress string, market map[string]interface{}) error {
	priceUSD := toFloat64(market["priceUsd"])
	liquidityUSD := toFloat64(getNested(market, "liquidity", "usd"))
	volumeH1 := toFloat64(getNested(market, "volume", "h1"))
	buysH1 := int(toFloat64(getNested(market, "txns", "h1", "buys")))
	sellsH1 := int(toFloat64(getNested(market, "txns", "h1", "sells")))
	raw, _ := json.Marshal(market)

	prevSnapshots, err := s.repo.GetLatestMarketSnapshots(ctx, tokenAddress, 1)
	if err != nil {
		return err
	}
	if err := s.repo.CreateMarketSnapshot(ctx, &model.TokenMarketSnapshot{
		TokenAddress: tokenAddress,
		PairAddress:  pairAddress,
		PriceUSD:     priceUSD,
		LiquidityUSD: liquidityUSD,
		VolumeH1:     volumeH1,
		BuysH1:       buysH1,
		SellsH1:      sellsH1,
		Raw:          raw,
	}); err != nil {
		return err
	}

	if len(prevSnapshots) == 0 || prevSnapshots[0].LiquidityUSD <= 0 || liquidityUSD <= 0 {
		return nil
	}

	change := (liquidityUSD - prevSnapshots[0].LiquidityUSD) / prevSnapshots[0].LiquidityUSD
	if change > -0.4 {
		return nil
	}

	details := map[string]interface{}{
		"prevLiquidityUsd": prevSnapshots[0].LiquidityUSD,
		"newLiquidityUsd":  liquidityUSD,
		"change":           change,
		"threshold":        -0.4,
	}
	detailsJSON, _ := json.Marshal(details)
	if err := s.repo.CreateTokenAlert(ctx, &model.TokenAlert{
		TokenAddress: tokenAddress,
		AlertType:    "LIQUIDITY_DROP",
		Severity:     "HIGH",
		Message:      "Liquidity dropped more than 40% within refresh window",
		Details:      detailsJSON,
	}); err != nil {
		return err
	}

	token, err := s.repo.GetTokenByAddress(ctx, tokenAddress)
	if err != nil {
		return err
	}
	alerts := make([]map[string]interface{}, 0)
	if len(token.MarketAlerts) > 0 {
		_ = json.Unmarshal(token.MarketAlerts, &alerts)
	}
	alerts = append(alerts, map[string]interface{}{
		"type":      "LIQUIDITY_DROP",
		"severity":  "HIGH",
		"change":    change,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	if len(alerts) > 20 {
		alerts = alerts[len(alerts)-20:]
	}
	alertsJSON, _ := json.Marshal(alerts)
	return s.repo.UpdateTokenAnalysis(ctx, tokenAddress, map[string]interface{}{"market_alerts": alertsJSON})
}

func (s *Scanner) logStatsLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := s.stats.snapshot()
			log.Printf("[Scanner] EnrichStats success=%d fail=%d refresh_success=%d refresh_fail=%d", stats.EnrichSuccess, stats.EnrichFailure, stats.RefreshSuccess, stats.RefreshFailure)
		}
	}
}

func normalizeGoPlus(data *basesdk.TokenSecurityData) map[string]interface{} {
	out := map[string]interface{}{
		"is_honeypot":             parseBinaryFlag(data.IsHoneypot),
		"buy_tax":                 parsePercentNumber(data.BuyTax),
		"sell_tax":                parsePercentNumber(data.SellTax),
		"is_mintable":             parseBinaryFlag(data.IsMintable),
		"can_take_back_ownership": parseBinaryFlag(data.CanTakeBackOwnership),
		"is_proxy":                parseBinaryFlag(data.IsProxy),
		"is_open_source":          parseBinaryFlag(data.IsOpenSource),
		"holder_count":            int(parsePlainNumber(data.HolderCount)),
		"lp_holder_count":         int(parsePlainNumber(data.LpHolderCount)),
		"creator_address":         strings.TrimSpace(data.CreatorAddress),
		"owner_address":           strings.TrimSpace(data.OwnerAddress),
		"total_supply":            strings.TrimSpace(data.TotalSupply),
	}
	return out
}

func normalizeDEXScreener(raw map[string]interface{}) map[string]interface{} {
	if raw == nil {
		raw = map[string]interface{}{}
	}
	out := map[string]interface{}{
		"raw":      raw,
		"priceUsd": asString(raw["priceUsd"]),
		"priceChange": map[string]interface{}{
			"m5":  toFloat64(getNested(raw, "priceChange", "m5")),
			"h1":  toFloat64(getNested(raw, "priceChange", "h1")),
			"h6":  toFloat64(getNested(raw, "priceChange", "h6")),
			"h24": toFloat64(getNested(raw, "priceChange", "h24")),
		},
		"volume": map[string]interface{}{
			"m5":  toFloat64(getNested(raw, "volume", "m5")),
			"h1":  toFloat64(getNested(raw, "volume", "h1")),
			"h6":  toFloat64(getNested(raw, "volume", "h6")),
			"h24": toFloat64(getNested(raw, "volume", "h24")),
		},
		"txns": map[string]interface{}{
			"m5": map[string]interface{}{
				"buys":  int(toFloat64(getNested(raw, "txns", "m5", "buys"))),
				"sells": int(toFloat64(getNested(raw, "txns", "m5", "sells"))),
			},
			"h1": map[string]interface{}{
				"buys":  int(toFloat64(getNested(raw, "txns", "h1", "buys"))),
				"sells": int(toFloat64(getNested(raw, "txns", "h1", "sells"))),
			},
			"h24": map[string]interface{}{
				"buys":  int(toFloat64(getNested(raw, "txns", "h24", "buys"))),
				"sells": int(toFloat64(getNested(raw, "txns", "h24", "sells"))),
			},
		},
		"liquidity": map[string]interface{}{
			"usd": toFloat64(getNested(raw, "liquidity", "usd")),
		},
	}
	return out
}

func parseBinaryFlag(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func parsePercentNumber(v string) float64 {
	clean := strings.TrimSpace(strings.TrimSuffix(v, "%"))
	if clean == "" {
		return 0
	}
	f, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0
	}
	if f < 0 {
		return 0
	}
	if f > 1 {
		f = f / 100
	}
	if f > 1 {
		return 1
	}
	return f
}

func parsePlainNumber(v string) float64 {
	clean := strings.TrimSpace(strings.TrimSuffix(v, "%"))
	if clean == "" {
		return 0
	}
	f, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0
	}
	if f < 0 {
		return 0
	}
	return f
}

func toFloat64(v interface{}) float64 {
	switch typed := v.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		if typed == "" {
			return 0
		}
		parsed, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func getNested(input map[string]interface{}, path ...string) interface{} {
	if input == nil {
		return nil
	}
	var current interface{} = input
	for _, key := range path {
		nextMap, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = nextMap[key]
	}
	return current
}

func asString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func trimErr(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) > 512 {
		return msg[:512]
	}
	return msg
}
