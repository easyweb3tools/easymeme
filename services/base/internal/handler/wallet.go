package handler

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"easyweb3/base/internal/model"
	"easyweb3/base/internal/repository"
	"easyweb3/base/internal/service"
	"easyweb3/base/pkg/ethereum"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	repo            *repository.Repository
	eth             *ethereum.Client
	walletMasterKey string
}

func NewWalletHandler(repo *repository.Repository, eth *ethereum.Client, walletMasterKey string) *WalletHandler {
	return &WalletHandler{repo: repo, eth: eth, walletMasterKey: strings.TrimSpace(walletMasterKey)}
}

type CreateWalletRequest struct {
	UserID string `json:"userId" binding:"required"`
}

type CreateWalletResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"userId"`
	Address string `json:"address"`
}

func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	serviceID, _ := c.Get("service_id")
	serviceIDStr, _ := serviceID.(string)

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Printf("generate key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate wallet"})
		return
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	encrypted, err := service.EncryptPrivateKey(h.walletMasterKey, privateKeyBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	wallet := &model.ManagedWallet{
		ServiceID:    serviceIDStr,
		UserID:       strings.TrimSpace(req.UserID),
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
	c.JSON(http.StatusOK, gin.H{"data": CreateWalletResponse{ID: wallet.ID, UserID: wallet.UserID, Address: wallet.Address}})
}

type WalletBalanceResponse struct {
	ID      string  `json:"id"`
	UserID  string  `json:"userId"`
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

func (h *WalletHandler) GetWalletBalance(c *gin.Context) {
	walletID := strings.TrimSpace(c.Param("walletId"))
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "walletId is required"})
		return
	}
	wallet, err := h.repo.GetManagedWalletByID(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	balanceWei, err := h.eth.GetBalance(c.Request.Context(), common.HexToAddress(wallet.Address))
	if err == nil {
		if balance, convErr := weiToBNB(balanceWei); convErr == nil {
			wallet.Balance = balance
			_ = h.repo.UpdateManagedWalletBalance(c.Request.Context(), wallet.ID, balance)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": WalletBalanceResponse{ID: wallet.ID, UserID: wallet.UserID, Address: wallet.Address, Balance: wallet.Balance}})
}

type ExecuteTradeRequest struct {
	Type         string `json:"type" binding:"required"`
	TokenAddress string `json:"tokenAddress" binding:"required"`
	AmountIn     string `json:"amountIn" binding:"required"`
	AmountOut    string `json:"amountOut"`
}

func (h *WalletHandler) ExecuteTrade(c *gin.Context) {
	walletID := strings.TrimSpace(c.Param("walletId"))
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "walletId is required"})
		return
	}

	var req ExecuteTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	wallet, err := h.repo.GetManagedWalletByID(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	privateKey, err := service.DecryptPrivateKey(h.walletMasterKey, wallet.EncryptedKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt key"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()

	tokenAddr := common.HexToAddress(req.TokenAddress)
	walletAddr := common.HexToAddress(wallet.Address)

	amountInWei, err := parseAmountToWei(req.AmountIn, req.Type, h.eth, tokenAddr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amountIn"})
		return
	}
	amountOutWei, _ := parseAmountToWei(req.AmountOut, req.Type, h.eth, tokenAddr)

	var txHash common.Hash
	switch strings.ToUpper(strings.TrimSpace(req.Type)) {
	case "BUY":
		txHash, err = h.eth.SwapExactETHForTokens(ctx, privateKey, tokenAddr, amountInWei, amountOutWei)
	case "SELL":
		approveAmount := new(big.Int).Mul(amountInWei, big.NewInt(2))
		_, _ = h.eth.ApproveToken(ctx, privateKey, tokenAddr, common.HexToAddress(ethereum.PancakeRouterV2), approveAmount)
		txHash, err = h.eth.SwapExactTokensForETH(ctx, privateKey, tokenAddr, amountInWei, amountOutWei)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trade type"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "trade failed"})
		return
	}

	receipt, _ := waitForReceipt(ctx, h.eth, txHash)
	status := "pending"
	if receipt != nil {
		if receipt.Status == 1 {
			status = "success"
		} else {
			status = "failed"
		}
	}

	if balanceWei, balErr := h.eth.GetBalance(ctx, walletAddr); balErr == nil {
		if balance, convErr := weiToBNB(balanceWei); convErr == nil {
			_ = h.repo.UpdateManagedWalletBalance(ctx, wallet.ID, balance)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"tx_hash": txHash.Hex(), "status": status}})
}

func parseAmountToWei(amount string, tradeType string, eth *ethereum.Client, tokenAddress common.Address) (*big.Int, error) {
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return big.NewInt(0), nil
	}
	d, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, err
	}
	if d <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}
	decimals := int32(18)
	if strings.EqualFold(tradeType, "SELL") {
		_, _, tokenDecimals, err := eth.GetTokenInfo(context.Background(), tokenAddress)
		if err == nil {
			decimals = int32(tokenDecimals)
		}
	}
	return decimalToWei(amount, decimals)
}

func decimalToWei(amount string, decimals int32) (*big.Int, error) {
	parts := strings.SplitN(amount, ".", 2)
	whole := parts[0]
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > int(decimals) {
		frac = frac[:decimals]
	}
	for len(frac) < int(decimals) {
		frac += "0"
	}
	combined := whole + frac
	combined = strings.TrimLeft(combined, "0")
	if combined == "" {
		combined = "0"
	}
	v, ok := new(big.Int).SetString(combined, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount")
	}
	return v, nil
}

func weiToBNB(wei *big.Int) (float64, error) {
	if wei == nil {
		return 0, fmt.Errorf("nil amount")
	}
	f := new(big.Float).SetInt(wei)
	div := new(big.Float).SetFloat64(1e18)
	res, _ := new(big.Float).Quo(f, div).Float64()
	return res, nil
}

func waitForReceipt(ctx context.Context, eth *ethereum.Client, txHash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	deadline := time.After(40 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("receipt timeout")
		case <-ticker.C:
			receipt, err := eth.TransactionReceipt(ctx, txHash)
			if err != nil {
				continue
			}
			if receipt != nil {
				return receipt, nil
			}
		}
	}
}

func privateKeyAddress(pk *ecdsa.PrivateKey) string {
	if pk == nil {
		return ""
	}
	return crypto.PubkeyToAddress(pk.PublicKey).Hex()
}
