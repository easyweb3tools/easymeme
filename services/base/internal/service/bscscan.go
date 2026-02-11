package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var scanEndpoints = map[string]string{
	"bsc":      "https://api.bscscan.com/api",
	"ethereum": "https://api.etherscan.io/api",
	"arbitrum": "https://api.arbiscan.io/api",
}

type BscScanClient struct {
	httpClient *http.Client
	apiKey     string
	mu         sync.Mutex
	lastCall   time.Time
}

type HolderDistribution struct {
	TopHolders []map[string]interface{} `json:"topHolders"`
	Top10Share float64                  `json:"top10Share"`
	Total      int                      `json:"total"`
	Source     string                   `json:"source"`
}

type CreatorHistory struct {
	CreatorAddress   string                   `json:"creatorAddress"`
	ContractAddress  string                   `json:"contractAddress"`
	CreationTxHash   string                   `json:"creationTxHash"`
	CreatedContracts []string                 `json:"createdContracts"`
	RecentTxs        []map[string]interface{} `json:"recentTxs"`
	Source           string                   `json:"source"`
}

func NewBscScanClient(apiKey string) *BscScanClient {
	return &BscScanClient{httpClient: &http.Client{Timeout: 10 * time.Second}, apiKey: strings.TrimSpace(apiKey)}
}

func (c *BscScanClient) FetchHolderDistribution(ctx context.Context, chain, tokenAddress string) (*HolderDistribution, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	query := url.Values{}
	query.Set("module", "token")
	query.Set("action", "tokenholderlist")
	query.Set("contractaddress", tokenAddress)
	query.Set("page", "1")
	query.Set("offset", "50")
	if c.apiKey != "" {
		query.Set("apikey", c.apiKey)
	}
	payload, err := c.getJSON(ctx, chain, query)
	if err != nil {
		return nil, err
	}
	result, _ := payload["result"].([]interface{})
	holders := make([]map[string]interface{}, 0, len(result))
	quantities := make([]float64, 0, len(result))
	for _, item := range result {
		rec, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		address, _ := rec["TokenHolderAddress"].(string)
		quantity := toFloat(rec["TokenHolderQuantity"])
		holders = append(holders, map[string]interface{}{"address": address, "quantity": quantity})
		quantities = append(quantities, quantity)
	}
	var totalQty float64
	for _, q := range quantities {
		totalQty += q
	}
	for i := range holders {
		if totalQty > 0 {
			holders[i]["share"] = toFloat(holders[i]["quantity"]) / totalQty
		} else {
			holders[i]["share"] = 0.0
		}
	}
	sort.Slice(holders, func(i, j int) bool { return toFloat(holders[i]["share"]) > toFloat(holders[j]["share"]) })
	var top10 float64
	for i := 0; i < len(holders) && i < 10; i++ {
		top10 += toFloat(holders[i]["share"])
	}
	return &HolderDistribution{TopHolders: holders, Top10Share: top10, Total: len(holders), Source: "scan"}, nil
}

func (c *BscScanClient) FetchCreatorHistory(ctx context.Context, chain, contractAddress string) (*CreatorHistory, error) {
	if contractAddress == "" {
		return nil, fmt.Errorf("contract address is required")
	}
	creatorAddress, txHash, err := c.fetchContractCreation(ctx, chain, contractAddress)
	if err != nil {
		return nil, err
	}
	recentTxs, createdContracts, err := c.fetchAddressTxSummary(ctx, chain, creatorAddress)
	if err != nil {
		return nil, err
	}
	return &CreatorHistory{CreatorAddress: creatorAddress, ContractAddress: contractAddress, CreationTxHash: txHash, CreatedContracts: createdContracts, RecentTxs: recentTxs, Source: "scan"}, nil
}

func (c *BscScanClient) fetchContractCreation(ctx context.Context, chain, contractAddress string) (string, string, error) {
	query := url.Values{}
	query.Set("module", "contract")
	query.Set("action", "getcontractcreation")
	query.Set("contractaddresses", contractAddress)
	if c.apiKey != "" {
		query.Set("apikey", c.apiKey)
	}
	payload, err := c.getJSON(ctx, chain, query)
	if err != nil {
		return "", "", err
	}
	items, _ := payload["result"].([]interface{})
	if len(items) == 0 {
		return "", "", fmt.Errorf("empty creation info")
	}
	first, _ := items[0].(map[string]interface{})
	creator, _ := first["contractCreator"].(string)
	if creator == "" {
		creator, _ = first["creatorAddress"].(string)
	}
	txHash, _ := first["txHash"].(string)
	if creator == "" {
		return "", "", fmt.Errorf("missing creator address")
	}
	return creator, txHash, nil
}

func (c *BscScanClient) fetchAddressTxSummary(ctx context.Context, chain, address string) ([]map[string]interface{}, []string, error) {
	if address == "" {
		return nil, nil, fmt.Errorf("address is required")
	}
	query := url.Values{}
	query.Set("module", "account")
	query.Set("action", "txlist")
	query.Set("address", address)
	query.Set("page", "1")
	query.Set("offset", "30")
	query.Set("sort", "desc")
	if c.apiKey != "" {
		query.Set("apikey", c.apiKey)
	}
	payload, err := c.getJSON(ctx, chain, query)
	if err != nil {
		return nil, nil, err
	}
	items, _ := payload["result"].([]interface{})
	recent := make([]map[string]interface{}, 0, len(items))
	contractSet := map[string]struct{}{}
	for _, item := range items {
		tx, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		recent = append(recent, map[string]interface{}{
			"hash":      str(tx["hash"]),
			"from":      str(tx["from"]),
			"to":        str(tx["to"]),
			"value":     str(tx["value"]),
			"timestamp": str(tx["timeStamp"]),
		})
		if contractAddr := str(tx["contractAddress"]); contractAddr != "" {
			contractSet[strings.ToLower(contractAddr)] = struct{}{}
		}
	}
	contracts := make([]string, 0, len(contractSet))
	for contract := range contractSet {
		contracts = append(contracts, contract)
	}
	sort.Strings(contracts)
	if len(contracts) > 20 {
		contracts = contracts[:20]
	}
	if len(recent) > 20 {
		recent = recent[:20]
	}
	return recent, contracts, nil
}

func (c *BscScanClient) getJSON(ctx context.Context, chain string, query url.Values) (map[string]interface{}, error) {
	c.waitRateLimit()
	endpoint, ok := scanEndpoints[strings.ToLower(strings.TrimSpace(chain))]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build bscscan request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request bscscan: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("bscscan status %d: %s", resp.StatusCode, string(body))
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode bscscan response: %w", err)
	}
	status, _ := payload["status"].(string)
	msg := strings.ToLower(str(payload["message"]))
	if status == "0" && !strings.Contains(msg, "no transactions") {
		return nil, fmt.Errorf("bscscan error: %s", str(payload["result"]))
	}
	return payload, nil
}

func (c *BscScanClient) waitRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastCall.IsZero() {
		c.lastCall = time.Now()
		return
	}
	interval := 220 * time.Millisecond
	if c.apiKey == "" {
		interval = time.Second
	}
	wait := interval - time.Since(c.lastCall)
	if wait > 0 {
		time.Sleep(wait)
	}
	c.lastCall = time.Now()
}

func str(v interface{}) string {
	s, _ := v.(string)
	return s
}

func toFloat(v interface{}) float64 {
	s := str(v)
	if s == "" {
		if f, ok := v.(float64); ok {
			return f
		}
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
