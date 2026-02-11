package basesdk

import (
	"context"
	"fmt"
)

type TokenSecurityData struct {
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
	Raw                  map[string]interface{} `json:"raw,omitempty"`
}

func (c *Client) GetTokenSecurity(ctx context.Context, chain, address string) (*TokenSecurityData, error) {
	var result TokenSecurityData
	path := fmt.Sprintf("/api/v1/tokens/%s/%s/security", chain, address)
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
