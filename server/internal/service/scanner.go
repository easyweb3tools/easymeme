package service

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"strings"
	"time"

	"easymeme/internal/model"
	"easymeme/internal/repository"
	"easymeme/pkg/ethereum"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

type Scanner struct {
	client *ethereum.Client
	repo   *repository.Repository
	hub    Broadcaster
}

type Broadcaster interface {
	Broadcast(payload interface{})
}

func NewScanner(client *ethereum.Client, repo *repository.Repository, hub Broadcaster) *Scanner {
	return &Scanner{
		client: client,
		repo:   repo,
		hub:    hub,
	}
}

func (s *Scanner) Start(ctx context.Context) error {
	logs, sub, err := s.client.SubscribePairCreated(ctx)
	if err != nil {
		log.Printf("[Scanner] Subscription unavailable: %v", err)
		go s.pollPairCreated(ctx)
		return nil
	}

	log.Println("[Scanner] Started listening for PairCreated events...")

	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Printf("[Scanner] Subscription error: %v", err)
				return
			case vLog := <-logs:
				go s.handlePairCreated(ctx, vLog)
			case <-ctx.Done():
				log.Println("[Scanner] Stopping...")
				return
			}
		}
	}()

	return nil
}

func (s *Scanner) pollPairCreated(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var lastBlock uint64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest, err := s.client.LatestBlockNumber(ctx)
			if err != nil {
				log.Printf("[Scanner] BlockNumber error: %v", err)
				continue
			}

			var from uint64
			if lastBlock == 0 {
				if latest > 5000 {
					from = latest - 5000
				} else {
					from = 0
				}
			} else {
				from = lastBlock + 1
			}

			if from > latest {
				continue
			}

			logs, err := s.client.GetPairCreatedLogs(ctx, from, latest)
			if err != nil {
				log.Printf("[Scanner] Poll logs error: %v", err)
				continue
			}

			for _, vLog := range logs {
				s.handlePairCreated(ctx, vLog)
			}

			lastBlock = latest
		}
	}
}

func (s *Scanner) handlePairCreated(ctx context.Context, vLog types.Log) {
	token0 := common.HexToAddress(vLog.Topics[1].Hex())
	token1 := common.HexToAddress(vLog.Topics[2].Hex())
	if len(vLog.Data) < 32 {
		return
	}
	pairAddr := common.BytesToAddress(vLog.Data[:32])

	wbnb := common.HexToAddress(ethereum.WBNB)
	var targetToken common.Address
	if strings.EqualFold(token0.Hex(), wbnb.Hex()) {
		targetToken = token1
	} else if strings.EqualFold(token1.Hex(), wbnb.Hex()) {
		targetToken = token0
	} else {
		return
	}

	log.Printf("[Scanner] New pair: %s, Token: %s", pairAddr.Hex(), targetToken.Hex())

	if s.repo.TokenExists(ctx, targetToken.Hex()) {
		return
	}

	name, symbol, decimals, _ := s.client.GetTokenInfo(ctx, targetToken)

	reserve0, reserve1, _ := s.client.GetPairReserves(ctx, pairAddr)
	var liquidity *big.Int
	if strings.EqualFold(token0.Hex(), wbnb.Hex()) {
		liquidity = reserve0
	} else {
		liquidity = reserve1
	}
	if liquidity == nil {
		liquidity = big.NewInt(0)
	}

	token := &model.Token{
		Address:          targetToken.Hex(),
		Name:             name,
		Symbol:           symbol,
		Decimals:         int(decimals),
		PairAddress:      pairAddr.Hex(),
		Dex:              "pancakeswap",
		InitialLiquidity: decimal.NewFromBigInt(liquidity, -18),
	}

	token.RiskScore = 0
	token.RiskLevel = "pending"
	token.AnalysisStatus = "pending"
	token.IsGoldenDog = false
	token.IsHoneypot = false
	token.BuyTax = 0
	token.SellTax = 0
	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"status": "pending",
	})
	token.RiskDetails = detailsJSON

	if err := s.repo.CreateToken(ctx, token); err != nil {
		log.Printf("[Scanner] Save token error: %v", err)
		return
	}

	s.hub.Broadcast(map[string]interface{}{
		"type":  "new_token",
		"token": token,
	})

	log.Printf("[Scanner] Token saved: %s (%s), Risk: %d", symbol, targetToken.Hex(), token.RiskScore)
}
