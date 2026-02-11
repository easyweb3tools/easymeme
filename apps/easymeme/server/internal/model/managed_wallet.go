package model

import "time"

type ManagedWallet struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"index;not null" json:"user_id"`
	Address      string    `gorm:"uniqueIndex;not null" json:"address"`
	EncryptedKey []byte    `json:"-"`
	Balance      float64   `gorm:"default:0" json:"balance"`
	MaxBalance   float64   `gorm:"default:5" json:"max_balance"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ManagedWallet) TableName() string {
	return "managed_wallets"
}
