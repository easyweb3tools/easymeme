package model

import "time"

type ServiceCredential struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceID string    `gorm:"size:50;uniqueIndex;not null" json:"service_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ServiceCredential) TableName() string { return "service_credentials" }
