package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Trade struct {
	ID           string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserAddress  string          `gorm:"index;not null" json:"user_address"`
	TokenAddress string          `gorm:"index;not null" json:"token_address"`
	TokenSymbol  string          `json:"token_symbol"`
	Type         string          `json:"type"` // buy, sell
	AmountIn     decimal.Decimal `gorm:"type:decimal(36,18)" json:"amount_in"`
	AmountOut    decimal.Decimal `gorm:"type:decimal(36,18)" json:"amount_out"`
	TxHash       string          `gorm:"uniqueIndex" json:"tx_hash"`
	Status       string          `json:"status"` // pending, success, failed
	GasUsed      decimal.Decimal `gorm:"type:decimal(36,18)" json:"gas_used"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

func (Trade) TableName() string {
	return "trades"
}
