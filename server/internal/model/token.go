package model

import (
    "time"

    "github.com/shopspring/decimal"
    "gorm.io/datatypes"
)

type Token struct {
    ID               string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Address          string          `gorm:"uniqueIndex;not null" json:"address"`
    Name             string          `json:"name"`
    Symbol           string          `json:"symbol"`
    Decimals         int             `gorm:"default:18" json:"decimals"`
    PairAddress      string          `json:"pair_address"`
    Dex              string          `gorm:"default:pancakeswap" json:"dex"`
    InitialLiquidity decimal.Decimal `gorm:"type:decimal(36,18)" json:"initial_liquidity"`
    RiskScore        int             `json:"risk_score"`
    RiskLevel        string          `json:"risk_level"` // safe, warning, danger
    RiskDetails      datatypes.JSON  `json:"risk_details"`
    IsHoneypot       bool            `gorm:"default:false" json:"is_honeypot"`
    BuyTax           float64         `json:"buy_tax"`
    SellTax          float64         `json:"sell_tax"`
    CreatorAddress   string          `json:"creator_address"`
    CreatedAt        time.Time       `gorm:"autoCreateTime;index" json:"created_at"`
    UpdatedAt        time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Token) TableName() string {
    return "tokens"
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
