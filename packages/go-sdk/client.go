package basesdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, serviceToken string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:      strings.TrimSpace(serviceToken),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("base-service %s %d: %s", req.URL.Path, resp.StatusCode, string(body))
	}
	if result != nil {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return json.Unmarshal(body, result)
		}
		if len(envelope.Data) == 0 {
			return nil
		}
		return json.Unmarshal(envelope.Data, result)
	}
	return nil
}
