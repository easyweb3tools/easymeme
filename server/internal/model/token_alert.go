package model

import (
	"time"

	"gorm.io/datatypes"
)

type TokenAlert struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TokenAddress string         `gorm:"index;not null" json:"token_address"`
	AlertType    string         `gorm:"index;not null" json:"alert_type"`
	Severity     string         `json:"severity"`
	Message      string         `json:"message"`
	Details      datatypes.JSON `json:"details"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
}

func (TokenAlert) TableName() string {
	return "token_alerts"
}
