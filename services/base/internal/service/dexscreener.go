package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type DEXScreenerClient struct {
	httpClient *http.Client
}

func NewDEXScreenerClient() *DEXScreenerClient {
	return &DEXScreenerClient{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (c *DEXScreenerClient) GetPairData(ctx context.Context, chain, pairAddress string) (map[string]interface{}, error) {
	if pairAddress == "" {
		return nil, fmt.Errorf("pair address is required")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain is required")
	}
	endpoint := fmt.Sprintf("https://api.dexscreener.com/latest/dex/pairs/%s/%s", strings.ToLower(chain), strings.ToLower(pairAddress))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build dexscreener request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request dexscreener: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("dexscreener status %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Pairs []map[string]interface{} `json:"pairs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode dexscreener response: %w", err)
	}
	if len(payload.Pairs) == 0 {
		return nil, fmt.Errorf("dexscreener empty pairs for %s", pairAddress)
	}
	return payload.Pairs[0], nil
}
