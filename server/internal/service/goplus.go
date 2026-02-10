package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const goPlusSecurityEndpoint = "https://api.gopluslabs.io/api/v1/token_security/56"

type GoPlusSecurityData struct {
	IsHoneypot           string                 `json:"is_honeypot"`
	BuyTax               string                 `json:"buy_tax"`
	SellTax              string                 `json:"sell_tax"`
	IsMintable           string                 `json:"is_mintable"`
	CanTakeBackOwnership string                 `json:"can_take_back_ownership"`
	IsProxy              string                 `json:"is_proxy"`
	IsOpenSource         string                 `json:"is_open_source"`
	HolderCount          string                 `json:"holder_count"`
	LpHolderCount        string                 `json:"lp_holder_count"`
	CreatorAddress       string                 `json:"creator_address"`
	OwnerAddress         string                 `json:"owner_address"`
	TotalSupply          string                 `json:"total_supply"`
	Raw                  map[string]interface{} `json:"-"`
}

type GoPlusClient struct {
	httpClient *http.Client
	mu         sync.Mutex
	lastCall   time.Time
}

func NewGoPlusClient() *GoPlusClient {
	return &GoPlusClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *GoPlusClient) GetTokenSecurity(ctx context.Context, tokenAddress string) (*GoPlusSecurityData, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}

	c.waitRateLimit()

	query := url.Values{}
	query.Set("contract_addresses", tokenAddress)
	endpoint := goPlusSecurityEndpoint + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build goplus request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request goplus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("goplus status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Code    int                        `json:"code"`
		Message string                     `json:"message"`
		Result  map[string]json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode goplus response: %w", err)
	}
	if len(payload.Result) == 0 {
		return nil, fmt.Errorf("goplus empty result for %s", tokenAddress)
	}

	var raw json.RawMessage
	for key, value := range payload.Result {
		if strings.EqualFold(key, tokenAddress) {
			raw = value
			break
		}
	}
	if len(raw) == 0 {
		for _, value := range payload.Result {
			raw = value
			break
		}
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("goplus missing token payload for %s", tokenAddress)
	}

	var item GoPlusSecurityData
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, fmt.Errorf("decode goplus token payload: %w", err)
	}
	if err := json.Unmarshal(raw, &item.Raw); err != nil {
		item.Raw = map[string]interface{}{}
	}

	return &item, nil
}

func (c *GoPlusClient) waitRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lastCall.IsZero() {
		c.lastCall = time.Now()
		return
	}
	elapsed := time.Since(c.lastCall)
	wait := 2*time.Second - elapsed
	if wait > 0 {
		time.Sleep(wait)
	}
	c.lastCall = time.Now()
}
