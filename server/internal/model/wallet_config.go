package model

import (
	"time"

	"gorm.io/datatypes"
)

type WalletConfig struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string         `gorm:"uniqueIndex;not null" json:"user_id"`
	Config    datatypes.JSON `json:"config"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (WalletConfig) TableName() string {
	return "wallet_configs"
}
