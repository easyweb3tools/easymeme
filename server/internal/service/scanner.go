package service

import (
    "context"
    "encoding/json"
    "log"
    "math/big"
    "strings"

    "easymeme/internal/model"
    "easymeme/internal/repository"
    "easymeme/pkg/ethereum"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/shopspring/decimal"
)

type Scanner struct {
    client   *ethereum.Client
    repo     *repository.Repository
    analyzer *Analyzer
    hub      Broadcaster
}

type Broadcaster interface {
    Broadcast(payload interface{})
}

func NewScanner(client *ethereum.Client, repo *repository.Repository, analyzer *Analyzer, hub Broadcaster) *Scanner {
    return &Scanner{
        client:   client,
        repo:     repo,
        analyzer: analyzer,
        hub:      hub,
    }
}

func (s *Scanner) Start(ctx context.Context) error {
    logs, sub, err := s.client.SubscribePairCreated(ctx)
    if err != nil {
        log.Printf("[Scanner] Subscription unavailable: %v", err)
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

    riskResult := s.analyzer.Analyze(ctx, targetToken)
    token.RiskScore = riskResult.Score
    token.RiskLevel = string(riskResult.Level)
    token.IsHoneypot = riskResult.IsHoneypot
    token.BuyTax = riskResult.BuyTax
    token.SellTax = riskResult.SellTax

    detailsJSON, _ := json.Marshal(riskResult.Details)
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
