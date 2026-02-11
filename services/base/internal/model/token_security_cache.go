package model

import "time"

type TokenSecurityCache struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Chain     string    `gorm:"size:20;not null;uniqueIndex:idx_security_chain_address" json:"chain"`
	Address   string    `gorm:"size:66;not null;uniqueIndex:idx_security_chain_address" json:"address"`
	Data      []byte    `gorm:"type:jsonb;not null" json:"data"`
	FetchedAt time.Time `gorm:"autoCreateTime" json:"fetched_at"`
}

func (TokenSecurityCache) TableName() string { return "token_security_cache" }
