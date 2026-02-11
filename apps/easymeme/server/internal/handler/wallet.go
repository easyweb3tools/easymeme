package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"easyweb3/apps/easymeme/internal/model"
	"easyweb3/apps/easymeme/internal/repository"
	"easyweb3/apps/easymeme/pkg/ethereum"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type WalletHandler struct {
	repo *repository.Repository
	eth  *ethereum.Client
}

func NewWalletHandler(repo *repository.Repository, eth *ethereum.Client) *WalletHandler {
	return &WalletHandler{repo: repo, eth: eth}
}

type CreateWalletRequest struct {
	UserID string `json:"userId"`
}

type CreateWalletResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"userId"`
	Address string `json:"address"`
}

// CreateWallet godoc
// @Summary Create managed wallet
// @Description Create a managed wallet and store encrypted private key
// @Tags wallet
// @Param payload body CreateWalletRequest true "Create wallet payload"
// @Success 200 {object} map[string]CreateWalletResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallet/create [post]
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Printf("generate key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate wallet"})
		return
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	encrypted, err := encryptPrivateKey(privateKeyBytes)
	if err != nil {
		if err == ErrMissingMasterKey {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "WALLET_MASTER_KEY is required"})
			return
		}
		log.Printf("encrypt key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt key"})
		return
	}

	wallet := &model.ManagedWallet{
		UserID:       req.UserID,
		Address:      address,
		EncryptedKey: encrypted,
		Balance:      0,
		MaxBalance:   5,
	}
	if err := h.repo.CreateManagedWallet(c.Request.Context(), wallet); err != nil {
		log.Printf("create wallet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wallet"})
		return
	}

	resp := CreateWalletResponse{
		ID:      wallet.ID,
		UserID:  wallet.UserID,
		Address: wallet.Address,
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

type WalletBalanceResponse struct {
	UserID  string  `json:"userId"`
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

// GetWalletBalance godoc
// @Summary Get managed wallet balance
// @Description Get managed wallet balance by user
// @Tags wallet
// @Param userId query string true "User ID"
// @Success 200 {object} map[string]WalletBalanceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallet/balance [get]
func (h *WalletHandler) GetWalletBalance(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	wallet, err := h.repo.GetManagedWalletByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": WalletBalanceResponse{
		UserID:  wallet.UserID,
		Address: wallet.Address,
		Balance: wallet.Balance,
	}})
}

// GetWalletInfo godoc
// @Summary Get managed wallet info
// @Description Get managed wallet address and balance (userId optional, fallback to EASYMEME_USER_ID)
// @Tags wallet
// @Param userId query string false "User ID"
// @Success 200 {object} map[string]WalletBalanceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallet/info [get]
func (h *WalletHandler) GetWalletInfo(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("userId"))
	if userID == "" {
		userID = strings.TrimSpace(os.Getenv("EASYMEME_USER_ID"))
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	wallet, err := h.repo.GetManagedWalletByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": WalletBalanceResponse{
		UserID:  wallet.UserID,
		Address: wallet.Address,
		Balance: wallet.Balance,
	}})
}

type WithdrawRequest struct {
	UserID string  `json:"userId"`
	Amount float64 `json:"amount"`
}

type AIPositionResponse struct {
	UserID       string `json:"user_id"`
	TokenAddress string `json:"token_address"`
	TokenSymbol  string `json:"token_symbol"`
	Quantity     string `json:"quantity"`
	CostBNB      string `json:"cost_bnb"`
	UpdatedAt    string `json:"updated_at"`
}

// GetAIPositions godoc
// @Summary Get AI positions
// @Description Get AI positions by user (userId optional, fallback to EASYMEME_USER_ID)
// @Tags ai-trades
// @Param userId query string false "User ID"
// @Success 200 {object} map[string][]AIPositionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/ai-positions [get]
func (h *WalletHandler) GetAIPositions(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("userId"))
	if userID == "" {
		userID = strings.TrimSpace(os.Getenv("EASYMEME_USER_ID"))
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	positions, err := h.repo.ListAIPositionsByUser(c.Request.Context(), userID)
	if err != nil {
		log.Printf("list ai positions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load positions"})
		return
	}

	resp := make([]AIPositionResponse, 0, len(positions))
	for _, pos := range positions {
		resp = append(resp, AIPositionResponse{
			UserID:       pos.UserID,
			TokenAddress: pos.TokenAddress,
			TokenSymbol:  pos.TokenSymbol,
			Quantity:     pos.Quantity.String(),
			CostBNB:      pos.CostBNB.String(),
			UpdatedAt:    pos.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

type WithdrawResponse struct {
	UserID  string  `json:"userId"`
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

// Withdraw godoc
// @Summary Withdraw from managed wallet
// @Description Decrease managed wallet balance (placeholder, no on-chain tx)
// @Tags wallet
// @Param payload body WithdrawRequest true "Withdraw payload"
// @Success 200 {object} map[string]WithdrawResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallet/withdraw [post]
func (h *WalletHandler) Withdraw(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 || strings.TrimSpace(req.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	wallet, err := h.repo.GetManagedWalletByUser(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	if req.Amount > wallet.Balance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
		return
	}

	newBalance := wallet.Balance - req.Amount
	if err := h.repo.UpdateManagedWalletBalance(c.Request.Context(), wallet.ID, newBalance); err != nil {
		log.Printf("update wallet balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": WithdrawResponse{
		UserID:  wallet.UserID,
		Address: wallet.Address,
		Balance: newBalance,
	}})
}

type WalletConfigRequest struct {
	UserID string                 `json:"userId"`
	Config map[string]interface{} `json:"config"`
}

// UpsertWalletConfig godoc
// @Summary Upsert wallet config
// @Description Upsert auto-trade config for managed wallet
// @Tags wallet
// @Param payload body WalletConfigRequest true "Wallet config payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallet/config [post]
func (h *WalletHandler) UpsertWalletConfig(c *gin.Context) {
	var req WalletConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Config == nil {
		req.Config = map[string]interface{}{}
	}
	payload, _ := json.Marshal(req.Config)
	if err := h.repo.UpsertWalletConfig(c.Request.Context(), req.UserID, payload); err != nil {
		log.Printf("upsert wallet config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type ExecuteTradeRequest struct {
	UserID       string  `json:"userId"`
	TokenAddress string  `json:"tokenAddress"`
	TokenSymbol  string  `json:"tokenSymbol"`
	Type         string  `json:"type"`     // BUY | SELL
	AmountIn     string  `json:"amountIn"` // BNB for BUY, Token for SELL
	AmountOut    string  `json:"amountOut"`
	Reason       string  `json:"decisionReason"`
	StrategyUsed string  `json:"strategyUsed"`
	GoldenScore  int     `json:"goldenDogScore"`
	ProfitLoss   float64 `json:"profitLoss"`
	Force        bool    `json:"force"`
}

// ExecuteTrade godoc
// @Summary Execute trade
// @Description Execute managed wallet trade on BSC
// @Tags wallet
// @Param payload body ExecuteTradeRequest true "Execute trade payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/wallet/execute-trade [post]
func (h *WalletHandler) ExecuteTrade(c *gin.Context) {
	var req ExecuteTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	userID := strings.TrimSpace(req.UserID)
	if userID == "" || req.TokenAddress == "" || req.Type == "" || req.AmountIn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	wallet, err := h.repo.GetManagedWalletByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	config, _ := h.loadWalletConfig(c.Request.Context(), userID)

	privateKey, err := decryptPrivateKey(wallet.EncryptedKey)
	if err != nil {
		log.Printf("decrypt key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt key"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	tokenAddr := common.HexToAddress(req.TokenAddress)
	walletAddr := common.HexToAddress(wallet.Address)
	preBNB, _ := h.eth.GetBalance(ctx, walletAddr)
	preToken, _ := h.eth.TokenBalance(ctx, tokenAddr, walletAddr)

	var txHash common.Hash
	switch strings.ToUpper(req.Type) {
	case "BUY":
		amountInWei, err := parseAmountToWei(req.AmountIn, req.Type, h.eth, tokenAddr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amountIn"})
			return
		}
		minOutWei, _ := parseAmountToWei(req.AmountOut, req.Type, h.eth, tokenAddr)
		if config.Enabled {
			if config.MinGoldenDogScore > 0 && req.GoldenScore < config.MinGoldenDogScore {
				c.JSON(http.StatusBadRequest, gin.H{"error": "golden dog score below threshold"})
				return
			}
			if config.MaxAmountPerTrade > 0 {
				if amountIn, err := decimal.NewFromString(req.AmountIn); err == nil {
					if amountIn.GreaterThan(decimal.NewFromFloat(config.MaxAmountPerTrade)) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "amount exceeds max per trade"})
						return
					}
				}
			}
			if config.DailyBudget > 0 {
				used, err := h.sumDailyBuyAmount(c.Request.Context(), userID)
				if err == nil {
					if current, err := decimal.NewFromString(req.AmountIn); err == nil {
						if used.Add(current).GreaterThan(decimal.NewFromFloat(config.DailyBudget)) {
							c.JSON(http.StatusBadRequest, gin.H{"error": "daily budget exceeded"})
							return
						}
					}
				}
			}
			if config.MaxDailyLoss > 0 {
				loss, err := h.sumDailyLoss(c.Request.Context(), userID)
				if err == nil && loss.GreaterThan(decimal.NewFromFloat(config.MaxDailyLoss)) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "max daily loss exceeded"})
					return
				}
			}
		}
		if wallet.MaxBalance > 0 {
			if amountIn, err := decimal.NewFromString(req.AmountIn); err == nil {
				if amountIn.GreaterThan(decimal.NewFromFloat(wallet.MaxBalance)) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "amount exceeds max balance limit"})
					return
				}
			}
		}
		if balanceWei, err := h.eth.GetBalance(ctx, walletAddr); err == nil {
			if balanceWei.Cmp(amountInWei) < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
				return
			}
		}
		txHash, err = h.eth.SwapExactETHForTokens(ctx, privateKey, tokenAddr, amountInWei, minOutWei)
	case "SELL":
		tokenBalance, balErr := h.eth.TokenBalance(ctx, tokenAddr, walletAddr)
		if balErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read token balance"})
			return
		}

		var amountInWei *big.Int
		var err error
		minOutWei, _ := parseAmountToWei(req.AmountOut, req.Type, h.eth, tokenAddr)
		_, _, decimals, _ := h.eth.GetTokenInfo(ctx, tokenAddr)

		if strings.EqualFold(req.AmountIn, "ALL") || strings.EqualFold(req.AmountIn, "100%") {
			amountInWei = tokenBalance
			req.AmountIn = formatAmount(amountInWei, int32(decimals))
		} else if ratio, ok := parseRatioAmount(req.AmountIn); ok {
			amountInWei = applyRatio(tokenBalance, ratio)
			req.AmountIn = formatAmount(amountInWei, int32(decimals))
		} else {
			amountInWei, err = parseAmountToWei(req.AmountIn, req.Type, h.eth, tokenAddr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amountIn"})
				return
			}
		}

		if config.Enabled && !req.Force {
			if config.StopLoss < 0 && req.ProfitLoss <= config.StopLoss {
				amountInWei = tokenBalance
				req.AmountIn = formatAmount(amountInWei, int32(decimals))
			} else if len(config.TakeProfitLevels) > 0 {
				levelIndex := matchedTakeProfitIndex(req.ProfitLoss, config.TakeProfitLevels)
				if levelIndex < 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "profit target not met"})
					return
				}
				ratio := pickTakeProfitRatio(levelIndex, config.TakeProfitAmounts)
				if ratio > 0 && ratio < 1 {
					amountInWei = applyRatio(tokenBalance, ratio)
					req.AmountIn = formatAmount(amountInWei, int32(decimals))
				}
			}
		}

		if tokenBalance.Cmp(amountInWei) < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient token balance"})
			return
		}

		approveAmount := new(big.Int).Mul(amountInWei, big.NewInt(2))
		_, _ = h.eth.ApproveToken(ctx, privateKey, tokenAddr, common.HexToAddress(ethereum.PancakeRouterV2), approveAmount)
		txHash, err = h.eth.SwapExactTokensForETH(ctx, privateKey, tokenAddr, amountInWei, minOutWei)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trade type"})
		return
	}
	if err != nil {
		log.Printf("execute trade: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "trade failed"})
		return
	}

	receipt, receiptErr := waitForReceipt(ctx, h.eth, txHash)
	status := "pending"
	errorMessage := ""
	var gasUsed string
	var blockNumber uint64
	if receiptErr == nil && receipt != nil {
		if receipt.Status == 1 {
			status = "success"
		} else {
			status = "failed"
		}
		gasUsed = strconv.FormatUint(receipt.GasUsed, 10)
		blockNumber = receipt.BlockNumber.Uint64()
	} else if receiptErr != nil {
		errorMessage = receiptErr.Error()
	}

	postBNB, _ := h.eth.GetBalance(ctx, walletAddr)
	postToken, _ := h.eth.TokenBalance(ctx, tokenAddr, walletAddr)
	amountOut := ""
	profitLoss := 0.0
	if strings.ToUpper(req.Type) == "BUY" {
		if postToken != nil && preToken != nil {
			if delta := new(big.Int).Sub(postToken, preToken); delta.Sign() > 0 {
				_, _, decimals, _ := h.eth.GetTokenInfo(ctx, tokenAddr)
				amountOut = formatAmount(delta, int32(decimals))
			}
		}
		h.upsertPositionAfterBuy(c.Request.Context(), userID, req.TokenAddress, req.TokenSymbol, req.AmountIn, amountOut)
	} else {
		if postBNB != nil && preBNB != nil {
			if delta := new(big.Int).Sub(postBNB, preBNB); delta.Sign() > 0 {
				amountOut = formatAmount(delta, 18)
			}
		}
		profitLoss = h.applyPositionAfterSell(c.Request.Context(), userID, req.TokenAddress, amountOut, req.AmountIn)
	}

	if balance, err := weiToBNB(postBNB); err == nil {
		_ = h.repo.UpdateManagedWalletBalance(c.Request.Context(), wallet.ID, balance)
	}

	aiTrade := &model.AITrade{
		UserID:         userID,
		TokenAddress:   req.TokenAddress,
		TokenSymbol:    req.TokenSymbol,
		Type:           strings.ToUpper(req.Type),
		AmountIn:       req.AmountIn,
		AmountOut:      amountOut,
		TxHash:         txHash.Hex(),
		Status:         status,
		GasUsed:        gasUsed,
		BlockNumber:    blockNumber,
		GoldenDogScore: req.GoldenScore,
		DecisionReason: req.Reason,
		StrategyUsed:   req.StrategyUsed,
		CurrentValue:   "",
		ProfitLoss:     profitLoss,
		ErrorMessage:   errorMessage,
	}
	_ = h.repo.CreateAITrade(c.Request.Context(), aiTrade)

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"tx_hash": txHash.Hex()}})
}

func encryptPrivateKey(privateKey []byte) ([]byte, error) {
	master := strings.TrimSpace(os.Getenv("WALLET_MASTER_KEY"))
	if master == "" {
		return nil, ErrMissingMasterKey
	}
	hash := sha256.Sum256([]byte(master))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, privateKey, nil)
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return []byte(hex.EncodeToString(out)), nil
}

var ErrMissingMasterKey = &errorString{s: "WALLET_MASTER_KEY is required"}

type errorString struct{ s string }

func (e *errorString) Error() string { return e.s }

type AutoTradeConfig struct {
	Enabled           bool      `json:"enabled"`
	MaxAmountPerTrade float64   `json:"maxAmountPerTrade"`
	MinGoldenDogScore int       `json:"minGoldenDogScore"`
	DailyBudget       float64   `json:"dailyBudget"`
	ConfirmThreshold  float64   `json:"confirmThreshold"`
	MaxDailyLoss      float64   `json:"maxDailyLoss"`
	TakeProfitLevels  []float64 `json:"takeProfitLevels"`
	TakeProfitAmounts []float64 `json:"takeProfitAmounts"`
	StopLoss          float64   `json:"stopLoss"`
}

func decryptPrivateKey(cipherHex []byte) (*ecdsa.PrivateKey, error) {
	master := strings.TrimSpace(os.Getenv("WALLET_MASTER_KEY"))
	if master == "" {
		return nil, ErrMissingMasterKey
	}
	raw, err := hex.DecodeString(string(cipherHex))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(master))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(plain)
}

func parseAmountToWei(amount string, tradeType string, client *ethereum.Client, tokenAddr common.Address) (*big.Int, error) {
	if strings.TrimSpace(amount) == "" {
		return big.NewInt(0), nil
	}
	decimals := int32(18)
	if strings.ToUpper(tradeType) == "SELL" {
		if client == nil {
			return nil, errors.New("missing client")
		}
		_, _, tokenDecimals, err := client.GetTokenInfo(context.Background(), tokenAddr)
		if err == nil {
			decimals = int32(tokenDecimals)
		}
	}
	value, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, err
	}
	scale := decimal.NewFromInt(1).Shift(decimals)
	wei := value.Mul(scale).BigInt()
	return wei, nil
}

func weiToBNB(value *big.Int) (float64, error) {
	if value == nil {
		return 0, errors.New("nil value")
	}
	dec := decimal.NewFromBigInt(value, 0)
	bnb := dec.Div(decimal.NewFromInt(1).Shift(18))
	f, _ := bnb.Float64()
	return f, nil
}

func formatAmount(value *big.Int, decimals int32) string {
	if value == nil {
		return ""
	}
	dec := decimal.NewFromBigInt(value, 0)
	out := dec.Div(decimal.NewFromInt(1).Shift(decimals))
	return out.String()
}

func waitForReceipt(ctx context.Context, client *ethereum.Client, hash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		receipt, err := client.Receipt(ctx, hash)
		if err == nil && receipt != nil {
			return receipt, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (h *WalletHandler) estimateProfitLoss(ctx context.Context, userID, tokenAddress, amountOut string) float64 {
	if amountOut == "" {
		return 0
	}
	lastBuy, err := h.repo.GetLatestAITradeByUserTokenType(ctx, userID, tokenAddress, "BUY")
	if err != nil || lastBuy == nil {
		return 0
	}
	buyIn, err := decimal.NewFromString(lastBuy.AmountIn)
	if err != nil || buyIn.Equal(decimal.Zero) {
		return 0
	}
	sellOut, err := decimal.NewFromString(amountOut)
	if err != nil {
		return 0
	}
	pl := sellOut.Sub(buyIn).Div(buyIn)
	f, _ := pl.Float64()
	return f
}

func (h *WalletHandler) upsertPositionAfterBuy(
	ctx context.Context,
	userID string,
	tokenAddress string,
	tokenSymbol string,
	amountInBNB string,
	amountOutToken string,
) {
	if amountOutToken == "" || amountInBNB == "" {
		return
	}
	buyCost, err := decimal.NewFromString(amountInBNB)
	if err != nil {
		return
	}
	buyQty, err := decimal.NewFromString(amountOutToken)
	if err != nil {
		return
	}
	if buyQty.LessThanOrEqual(decimal.Zero) {
		return
	}
	pos, err := h.repo.GetAIPosition(ctx, userID, tokenAddress)
	if err != nil || pos == nil {
		pos = &model.AIPosition{
			UserID:       userID,
			TokenAddress: tokenAddress,
			TokenSymbol:  tokenSymbol,
			Quantity:     buyQty,
			CostBNB:      buyCost,
		}
		_ = h.repo.UpsertAIPosition(ctx, pos)
		return
	}
	pos.Quantity = pos.Quantity.Add(buyQty)
	pos.CostBNB = pos.CostBNB.Add(buyCost)
	pos.TokenSymbol = tokenSymbol
	_ = h.repo.UpsertAIPosition(ctx, pos)
}

func (h *WalletHandler) applyPositionAfterSell(
	ctx context.Context,
	userID string,
	tokenAddress string,
	amountOutBNB string,
	amountInToken string,
) float64 {
	if amountOutBNB == "" || amountInToken == "" {
		return 0
	}
	sellOut, err := decimal.NewFromString(amountOutBNB)
	if err != nil {
		return 0
	}
	sellQty, err := decimal.NewFromString(amountInToken)
	if err != nil {
		return 0
	}
	if sellQty.LessThanOrEqual(decimal.Zero) {
		return 0
	}
	pos, err := h.repo.GetAIPosition(ctx, userID, tokenAddress)
	if err != nil || pos == nil || pos.Quantity.LessThanOrEqual(decimal.Zero) {
		return 0
	}
	avgCost := pos.CostBNB.Div(pos.Quantity)
	costSold := avgCost.Mul(sellQty)
	if costSold.LessThanOrEqual(decimal.Zero) {
		return 0
	}
	pl := sellOut.Sub(costSold).Div(costSold)
	f, _ := pl.Float64()

	pos.Quantity = pos.Quantity.Sub(sellQty)
	pos.CostBNB = pos.CostBNB.Sub(costSold)
	if pos.Quantity.LessThan(decimal.Zero) {
		pos.Quantity = decimal.Zero
		pos.CostBNB = decimal.Zero
	}
	_ = h.repo.UpsertAIPosition(ctx, pos)

	return f
}

func matchedTakeProfitIndex(value float64, levels []float64) int {
	index := -1
	for i, level := range levels {
		if value >= level {
			index = i
		}
	}
	return index
}

func pickTakeProfitRatio(index int, amounts []float64) float64 {
	if index < 0 {
		return 0
	}
	if len(amounts) == 0 {
		return 1
	}
	if index < len(amounts) {
		return amounts[index]
	}
	return amounts[len(amounts)-1]
}

func applyRatio(value *big.Int, ratio float64) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}
	dec := decimal.NewFromBigInt(value, 0)
	out := dec.Mul(decimal.NewFromFloat(ratio))
	return out.BigInt()
}

func parseRatioAmount(value string) (float64, bool) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, false
	}
	trimmed := strings.TrimSuffix(text, "%")
	ratio, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, false
	}
	if strings.HasSuffix(text, "%") {
		ratio = ratio / 100
	}
	if ratio <= 0 || ratio > 1 {
		return 0, false
	}
	return ratio, true
}

func (h *WalletHandler) loadWalletConfig(ctx context.Context, userID string) (AutoTradeConfig, error) {
	cfg := AutoTradeConfig{
		Enabled:           false,
		MinGoldenDogScore: 0,
	}
	record, err := h.repo.GetWalletConfig(ctx, userID)
	if err != nil || record == nil || len(record.Config) == 0 {
		return cfg, err
	}
	_ = json.Unmarshal(record.Config, &cfg)
	return cfg, nil
}

func (h *WalletHandler) sumDailyBuyAmount(ctx context.Context, userID string) (decimal.Decimal, error) {
	since := time.Now().Add(-24 * time.Hour)
	trades, err := h.repo.GetAITradesByUserSince(ctx, userID, since)
	if err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, t := range trades {
		if strings.ToUpper(t.Type) != "BUY" {
			continue
		}
		if v, err := decimal.NewFromString(t.AmountIn); err == nil {
			total = total.Add(v)
		}
	}
	return total, nil
}

func (h *WalletHandler) sumDailyLoss(ctx context.Context, userID string) (decimal.Decimal, error) {
	since := time.Now().Add(-24 * time.Hour)
	trades, err := h.repo.GetAITradesByUserSince(ctx, userID, since)
	if err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, t := range trades {
		if t.ProfitLoss < 0 {
			total = total.Add(decimal.NewFromFloat(-t.ProfitLoss))
		}
	}
	return total, nil
}

func profitMeetsTakeProfit(value float64, levels []float64) bool {
	for _, level := range levels {
		if value >= level {
			return true
		}
	}
	return false
}
