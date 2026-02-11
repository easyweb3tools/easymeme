package model

import "time"

type TokenPriceSnapshot struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TokenAddress string    `gorm:"index:idx_token_ts,priority:1;not null" json:"token_address"`
	TS           time.Time `gorm:"index:idx_token_ts,priority:2;not null" json:"ts"`
	PriceUSD     float64   `gorm:"not null" json:"price_usd"`
	LiquidityUSD float64   `json:"liquidity_usd"`
	Volume5mUSD  float64   `json:"volume_5m_usd"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TokenPriceSnapshot) TableName() string {
	return "token_price_snapshots"
}
