package basesdk

import (
	"context"
	"fmt"
)

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

func (c *Client) GetHolderDistribution(ctx context.Context, chain, address string) (*HolderDistribution, error) {
	var result HolderDistribution
	path := fmt.Sprintf("/api/v1/tokens/%s/%s/holders", chain, address)
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetCreatorHistory(ctx context.Context, chain, address string) (*CreatorHistory, error) {
	var result CreatorHistory
	path := fmt.Sprintf("/api/v1/tokens/%s/%s/creator", chain, address)
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
