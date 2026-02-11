package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AIPosition struct {
	ID           string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string          `gorm:"index;not null" json:"user_id"`
	TokenAddress string          `gorm:"index;not null" json:"token_address"`
	TokenSymbol  string          `json:"token_symbol"`
	Quantity     decimal.Decimal `gorm:"type:decimal(36,18)" json:"quantity"`
	CostBNB      decimal.Decimal `gorm:"type:decimal(36,18)" json:"cost_bnb"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AIPosition) TableName() string {
	return "ai_positions"
}
