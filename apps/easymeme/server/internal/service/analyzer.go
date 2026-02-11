package service

import (
	"context"
	"math/big"

	"easyweb3/apps/easymeme/pkg/ethereum"

	"github.com/ethereum/go-ethereum/common"
)

type RiskLevel string

const (
	RiskSafe    RiskLevel = "safe"
	RiskWarning RiskLevel = "warning"
	RiskDanger  RiskLevel = "danger"
)

type RiskDetails struct {
	CanMint           bool    `json:"can_mint"`
	CanPause          bool    `json:"can_pause"`
	CanBlacklist      bool    `json:"can_blacklist"`
	OwnerCanChangeTax bool    `json:"owner_can_change_tax"`
	LPLocked          bool    `json:"lp_locked"`
	ContractVerified  bool    `json:"contract_verified"`
	Top10Holding      float64 `json:"top10_holding"`
}

type RiskResult struct {
	Score      int         `json:"score"`
	Level      RiskLevel   `json:"level"`
	IsHoneypot bool        `json:"is_honeypot"`
	BuyTax     float64     `json:"buy_tax"`
	SellTax    float64     `json:"sell_tax"`
	Details    RiskDetails `json:"details"`
}

type Analyzer struct {
	client *ethereum.Client
}

func NewAnalyzer(client *ethereum.Client) *Analyzer {
	return &Analyzer{client: client}
}

func (a *Analyzer) Analyze(ctx context.Context, tokenAddr common.Address) RiskResult {
	details := RiskDetails{
		CanMint:           false,
		CanPause:          false,
		CanBlacklist:      false,
		OwnerCanChangeTax: false,
		LPLocked:          false,
		ContractVerified:  true,
		Top10Holding:      0,
	}

	isHoneypot := false
	if err := a.client.SimulateSell(ctx, tokenAddr, big.NewInt(1e18)); err != nil {
		isHoneypot = true
	}

	if isHoneypot {
		return RiskResult{
			Score:      0,
			Level:      RiskDanger,
			IsHoneypot: true,
			BuyTax:     0,
			SellTax:    0,
			Details:    details,
		}
	}

	score := 100
	level := RiskSafe
	if score < 40 {
		level = RiskDanger
	} else if score < 70 {
		level = RiskWarning
	}

	return RiskResult{
		Score:      score,
		Level:      level,
		IsHoneypot: false,
		BuyTax:     0,
		SellTax:    0,
		Details:    details,
	}
}
