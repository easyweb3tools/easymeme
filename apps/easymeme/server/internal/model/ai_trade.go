package model

import "time"

type AITrade struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"index;not null" json:"user_id"`
	TokenAddress string    `gorm:"index;not null" json:"token_address"`
	TokenSymbol  string    `json:"token_symbol"`
	Type         string    `json:"type"` // BUY | SELL
	AmountIn     string    `json:"amount_in"`
	AmountOut    string    `json:"amount_out"`
	TxHash       string    `gorm:"uniqueIndex" json:"tx_hash"`
	Timestamp    time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Status       string    `json:"status"` // pending | success | failed
	GasUsed      string    `json:"gas_used"`
	BlockNumber  uint64    `json:"block_number"`
	ErrorMessage string    `json:"error_message"`

	GoldenDogScore int    `json:"golden_dog_score"`
	DecisionReason string `json:"decision_reason"`
	StrategyUsed   string `json:"strategy_used"`

	CurrentValue string  `json:"current_value"`
	ProfitLoss   float64 `json:"profit_loss"`
}

func (AITrade) TableName() string {
	return "ai_trades"
}
