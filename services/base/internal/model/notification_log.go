package model

import "time"

type NotificationLog struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceID string    `gorm:"size:50;index;not null" json:"service_id"`
	Channel   string    `gorm:"size:20;not null" json:"channel"`
	Recipient string    `gorm:"size:200;not null" json:"recipient"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Status    string    `gorm:"size:20;default:sent" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (NotificationLog) TableName() string { return "notification_logs" }
