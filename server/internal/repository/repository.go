package repository

import (
	"context"
	"os"
	"time"

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
		db.AutoMigrate(
			&model.Token{},
			&model.Trade{},
			&model.ManagedWallet{},
			&model.WalletConfig{},
			&model.AITrade{},
			&model.AIPosition{},
			&model.TokenPriceSnapshot{},
		)
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

func (r *Repository) GetAnalyzedTokens(ctx context.Context, limit int) ([]model.Token, error) {
	var tokens []model.Token
	err := r.db.WithContext(ctx).
		Where("analysis_status = ?", "analyzed").
		Order("analyzed_at DESC").
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}

func (r *Repository) GetGoldenDogTokens(ctx context.Context, limit int) ([]model.Token, error) {
	var tokens []model.Token
	err := r.db.WithContext(ctx).
		Where("analysis_status = ?", "analyzed").
		Where("is_golden_dog = ?", true).
		Order("analyzed_at DESC").
		Limit(limit).
		Find(&tokens).Error
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

func (r *Repository) CreateManagedWallet(ctx context.Context, wallet *model.ManagedWallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *Repository) GetManagedWalletByUser(ctx context.Context, userID string) (*model.ManagedWallet, error) {
	var wallet model.ManagedWallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
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

func (r *Repository) UpsertWalletConfig(ctx context.Context, userID string, configJSON []byte) error {
	var existing model.WalletConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&existing).Error
	if err == nil {
		return r.db.WithContext(ctx).
			Model(&model.WalletConfig{}).
			Where("user_id = ?", userID).
			Update("config", configJSON).Error
	}
	return r.db.WithContext(ctx).Create(&model.WalletConfig{
		UserID: userID,
		Config: configJSON,
	}).Error
}

func (r *Repository) GetWalletConfig(ctx context.Context, userID string) (*model.WalletConfig, error) {
	var cfg model.WalletConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *Repository) GetAITrades(ctx context.Context, limit int) ([]model.AITrade, error) {
	var trades []model.AITrade
	err := r.db.WithContext(ctx).
		Order("timestamp DESC").
		Limit(limit).
		Find(&trades).Error
	return trades, err
}

func (r *Repository) GetAllAITrades(ctx context.Context) ([]model.AITrade, error) {
	var trades []model.AITrade
	err := r.db.WithContext(ctx).
		Order("timestamp DESC").
		Find(&trades).Error
	return trades, err
}

func (r *Repository) GetAITradesByUserSince(ctx context.Context, userID string, since time.Time) ([]model.AITrade, error) {
	var trades []model.AITrade
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("timestamp >= ?", since).
		Order("timestamp DESC").
		Find(&trades).Error
	return trades, err
}

func (r *Repository) GetLatestAITradeByUserTokenType(ctx context.Context, userID, tokenAddress, tradeType string) (*model.AITrade, error) {
	var trade model.AITrade
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("token_address = ?", tokenAddress).
		Where("type = ?", tradeType).
		Order("timestamp DESC").
		First(&trade).Error
	if err != nil {
		return nil, err
	}
	return &trade, nil
}

func (r *Repository) CreateAITrade(ctx context.Context, trade *model.AITrade) error {
	return r.db.WithContext(ctx).Create(trade).Error
}

func (r *Repository) GetAIPosition(ctx context.Context, userID, tokenAddress string) (*model.AIPosition, error) {
	var pos model.AIPosition
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("token_address = ?", tokenAddress).
		First(&pos).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *Repository) UpsertAIPosition(ctx context.Context, pos *model.AIPosition) error {
	if pos == nil {
		return nil
	}
	var existing model.AIPosition
	err := r.db.WithContext(ctx).
		Where("user_id = ?", pos.UserID).
		Where("token_address = ?", pos.TokenAddress).
		First(&existing).Error
	if err == nil {
		return r.db.WithContext(ctx).
			Model(&model.AIPosition{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"quantity":     pos.Quantity,
				"cost_bnb":     pos.CostBNB,
				"token_symbol": pos.TokenSymbol,
			}).Error
	}
	return r.db.WithContext(ctx).Create(pos).Error
}

func (r *Repository) ListAIPositionsByUser(ctx context.Context, userID string) ([]model.AIPosition, error) {
	var positions []model.AIPosition
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&positions).Error
	if err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *Repository) GetAITradeStats(ctx context.Context) (count int64, winRate float64, avgPL float64, err error) {
	var trades []model.AITrade
	if err = r.db.WithContext(ctx).Find(&trades).Error; err != nil {
		return 0, 0, 0, err
	}
	if len(trades) == 0 {
		return 0, 0, 0, nil
	}
	var wins int
	var totalPL float64
	for _, t := range trades {
		if t.ProfitLoss > 0 {
			wins++
		}
		totalPL += t.ProfitLoss
	}
	count = int64(len(trades))
	winRate = float64(wins) / float64(len(trades))
	avgPL = totalPL / float64(len(trades))
	return count, winRate, avgPL, nil
}

func (r *Repository) GetAnalyzedTokensSince(ctx context.Context, since time.Time, limit int, offset int) ([]model.Token, int64, error) {
	var tokens []model.Token
	var total int64
	q := r.db.WithContext(ctx).
		Model(&model.Token{}).
		Where("analysis_status = ?", "analyzed").
		Where("analyzed_at >= ?", since)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("analyzed_at DESC").Limit(limit).Offset(offset).Find(&tokens).Error; err != nil {
		return nil, 0, err
	}
	return tokens, total, nil
}

func (r *Repository) GetGoldenDogScoreDistributionSince(ctx context.Context, since time.Time, bucket int) (map[int]int64, int64, error) {
	if bucket <= 0 {
		bucket = 10
	}
	type row struct {
		Bucket int
		Count  int64
	}
	rows := []row{}
	var total int64

	q := r.db.WithContext(ctx).
		Model(&model.Token{}).
		Where("analysis_status = ?", "analyzed").
		Where("analyzed_at >= ?", since)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Select("FLOOR(golden_dog_score / ?)::int AS bucket, COUNT(*) AS count", bucket).
		Group("bucket").
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make(map[int]int64)
	for _, r := range rows {
		out[r.Bucket] = r.Count
	}
	return out, total, nil
}

func (r *Repository) UpsertTokenPriceSnapshot(ctx context.Context, s *model.TokenPriceSnapshot) error {
	if s == nil {
		return nil
	}
	var existing model.TokenPriceSnapshot
	err := r.db.WithContext(ctx).
		Where("token_address = ?", s.TokenAddress).
		Where("ts = ?", s.TS).
		First(&existing).Error
	if err == nil {
		return r.db.WithContext(ctx).Model(&model.TokenPriceSnapshot{}).
			Where("id = ?", existing.ID).
			Updates(map[string]any{
				"price_usd":      s.PriceUSD,
				"liquidity_usd":  s.LiquidityUSD,
				"volume_5m_usd":  s.Volume5mUSD,
			}).Error
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) GetTokenPriceSeries(ctx context.Context, tokenAddress string, from, to time.Time, limit int) ([]model.TokenPriceSnapshot, error) {
	var rows []model.TokenPriceSnapshot
	q := r.db.WithContext(ctx).
		Where("token_address = ?", tokenAddress).
		Where("ts >= ?", from).
		Where("ts <= ?", to).
		Order("ts ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
