package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type Token struct {
	ID                  string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Address             string          `gorm:"uniqueIndex;not null" json:"address"`
	Name                string          `json:"name"`
	Symbol              string          `json:"symbol"`
	Decimals            int             `gorm:"default:18" json:"decimals"`
	PairAddress         string          `json:"pair_address"`
	Dex                 string          `gorm:"default:pancakeswap" json:"dex"`
	InitialLiquidity    decimal.Decimal `gorm:"type:decimal(36,18)" json:"initial_liquidity"`
	AnalysisStatus      string          `gorm:"default:pending" json:"analysis_status"` // pending, enriching, enrich_failed, enriched, analyzed
	EnrichError         string          `json:"enrich_error"`
	EnrichAttempts      int             `gorm:"default:0" json:"enrich_attempts"`
	EnrichedAt          *time.Time      `json:"enriched_at"`
	RiskScore           int             `gorm:"default:0" json:"risk_score"`
	RiskLevel           string          `gorm:"default:pending" json:"risk_level"` // pending, safe, warning, danger
	RiskDetails         datatypes.JSON  `json:"risk_details"`
	MarketData          datatypes.JSON  `json:"market_data"`
	HolderData          datatypes.JSON  `json:"holder_data"`
	CreatorHistory      datatypes.JSON  `json:"creator_history"`
	MarketAlerts        datatypes.JSON  `json:"market_alerts"`
	SocialSignals       datatypes.JSON  `json:"social_signals"`
	SmartMoneySignals   datatypes.JSON  `json:"smart_money_signals"`
	LastMarketRefreshAt *time.Time      `json:"last_market_refresh_at"`
	AnalysisResult      datatypes.JSON  `json:"analysis_result"`
	IsGoldenDog         bool            `gorm:"default:false" json:"is_golden_dog"`
	GoldenDogScore      int             `gorm:"default:0" json:"golden_dog_score"`
	IsHoneypot          bool            `gorm:"default:false" json:"is_honeypot"`
	BuyTax              float64         `json:"buy_tax"`
	SellTax             float64         `json:"sell_tax"`
	CreatorAddress      string          `json:"creator_address"`
	CreatedAt           time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	AnalyzedAt          *time.Time      `json:"analyzed_at"`
}

func (Token) TableName() string {
	return "tokens"
}

func (t *Token) GoldenDogPhase() string {
	age := time.Since(t.CreatedAt)
	switch {
	case age <= 30*time.Minute:
		return "EARLY"
	case age <= 2*time.Hour:
		return "PEAK"
	case age <= 6*time.Hour:
		return "DECLINING"
	default:
		return "EXPIRED"
	}
}

func (t *Token) TimeDecayFactor() float64 {
	age := time.Since(t.CreatedAt)
	switch {
	case age <= 30*time.Minute:
		return 1.0
	case age <= 2*time.Hour:
		// Linear from 1.0 at 30m to 0.8 at 2h.
		progress := float64(age-30*time.Minute) / float64(90*time.Minute)
		return 1.0 - 0.2*progress
	case age <= 6*time.Hour:
		// Linear from 0.8 at 2h to 0.5 at 6h.
		progress := float64(age-2*time.Hour) / float64(4*time.Hour)
		return 0.8 - 0.3*progress
	default:
		return 0.4
	}
}

func (t *Token) EffectiveScore() int {
	score := float64(t.GoldenDogScore) * t.TimeDecayFactor()
	if score < 0 {
		return 0
	}
	return int(score + 0.5)
}

type RiskDetailsJSON struct {
	CanMint           bool    `json:"can_mint"`
	CanPause          bool    `json:"can_pause"`
	CanBlacklist      bool    `json:"can_blacklist"`
	OwnerCanChangeTax bool    `json:"owner_can_change_tax"`
	LPLocked          bool    `json:"lp_locked"`
	LPLockDays        int     `json:"lp_lock_days"`
	ContractVerified  bool    `json:"contract_verified"`
	Top10Holding      float64 `json:"top10_holding"`
}
