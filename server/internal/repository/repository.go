package repository

import (
	"context"
	"os"

	"easymeme/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func New(databaseURL string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	if os.Getenv("AUTO_MIGRATE") == "true" {
		db.AutoMigrate(&model.Token{}, &model.Trade{})
	}

	return &Repository{db: db}, nil
}

func (r *Repository) CreateToken(ctx context.Context, token *model.Token) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repository) GetTokenByAddress(ctx context.Context, address string) (*model.Token, error) {
	var token model.Token
	err := r.db.WithContext(ctx).Where("address = ?", address).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *Repository) GetLatestTokens(ctx context.Context, limit int) ([]model.Token, error) {
	var tokens []model.Token
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}

func (r *Repository) GetPendingTokens(ctx context.Context, limit int, minLiquidity float64) ([]model.Token, error) {
	var tokens []model.Token
	query := r.db.WithContext(ctx).
		Where("analysis_status = ? OR analysis_status IS NULL", "pending").
		Order("created_at DESC").
		Limit(limit)
	if minLiquidity > 0 {
		query = query.Where("initial_liquidity >= ?", minLiquidity)
	}
	err := query.Find(&tokens).Error
	return tokens, err
}

func (r *Repository) UpdateToken(ctx context.Context, token *model.Token) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *Repository) UpdateTokenAnalysis(ctx context.Context, address string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.Token{}).
		Where("address = ?", address).
		Updates(updates).Error
}

func (r *Repository) TokenExists(ctx context.Context, address string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.Token{}).Where("address = ?", address).Count(&count)
	return count > 0
}

func (r *Repository) CreateTrade(ctx context.Context, trade *model.Trade) error {
	return r.db.WithContext(ctx).Create(trade).Error
}

func (r *Repository) GetTradesByUser(ctx context.Context, userAddress string, limit int) ([]model.Trade, error) {
	var trades []model.Trade
	err := r.db.WithContext(ctx).
		Where("user_address = ?", userAddress).
		Order("created_at DESC").
		Limit(limit).
		Find(&trades).Error
	return trades, err
}

func (r *Repository) UpdateTradeStatus(ctx context.Context, txHash, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.Trade{}).
		Where("tx_hash = ?", txHash).
		Update("status", status).Error
}
