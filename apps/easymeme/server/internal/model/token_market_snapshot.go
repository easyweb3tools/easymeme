package model

import (
	"time"

	"gorm.io/datatypes"
)

type TokenMarketSnapshot struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TokenAddress string         `gorm:"index;not null" json:"token_address"`
	PairAddress  string         `gorm:"index" json:"pair_address"`
	PriceUSD     float64        `json:"price_usd"`
	LiquidityUSD float64        `json:"liquidity_usd"`
	VolumeH1     float64        `json:"volume_h1"`
	BuysH1       int            `json:"buys_h1"`
	SellsH1      int            `json:"sells_h1"`
	Raw          datatypes.JSON `json:"raw"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
}

func (TokenMarketSnapshot) TableName() string {
	return "token_market_snapshots"
}
