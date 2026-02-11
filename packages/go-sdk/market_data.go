package basesdk

import (
	"context"
	"fmt"
)

func (c *Client) GetPairData(ctx context.Context, chain, pairAddress string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/v1/market/%s/pairs/%s", chain, pairAddress)
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}
