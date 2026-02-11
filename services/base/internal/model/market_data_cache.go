package model

import "time"

type MarketDataCache struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Chain       string    `gorm:"size:20;not null;uniqueIndex:idx_market_chain_pair" json:"chain"`
	PairAddress string    `gorm:"size:66;not null;uniqueIndex:idx_market_chain_pair" json:"pair_address"`
	Data        []byte    `gorm:"type:jsonb;not null" json:"data"`
	FetchedAt   time.Time `gorm:"autoCreateTime" json:"fetched_at"`
}

func (MarketDataCache) TableName() string { return "market_data_cache" }
