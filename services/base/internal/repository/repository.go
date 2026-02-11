package repository

import (
	"context"
	"os"

	"easyweb3/base/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func New(databaseURL string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.AutoMigrate(
			&model.ManagedWallet{},
			&model.TokenSecurityCache{},
			&model.MarketDataCache{},
			&model.NotificationLog{},
			&model.ServiceCredential{},
		); err != nil {
			return nil, err
		}
	}
	return &Repository{db: db}, nil
}

func (r *Repository) CreateManagedWallet(ctx context.Context, wallet *model.ManagedWallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *Repository) GetManagedWalletByID(ctx context.Context, walletID string) (*model.ManagedWallet, error) {
	var wallet model.ManagedWallet
	if err := r.db.WithContext(ctx).Where("id = ?", walletID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *Repository) UpdateManagedWalletBalance(ctx context.Context, walletID string, balance float64) error {
	return r.db.WithContext(ctx).
		Model(&model.ManagedWallet{}).
		Where("id = ?", walletID).
		Update("balance", balance).Error
}

func (r *Repository) CreateNotificationLog(ctx context.Context, log *model.NotificationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
