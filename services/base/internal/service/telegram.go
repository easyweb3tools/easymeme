package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type TelegramClient struct {
	botToken   string
	httpClient *http.Client
}

func NewTelegramClient(botToken string) *TelegramClient {
	return &TelegramClient{
		botToken:   botToken,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TelegramClient) SendMessage(ctx context.Context, chatID, text, parseMode string) error {
	if c.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}
	if parseMode == "" {
		parseMode = "HTML"
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.botToken)
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", text)
	params.Set("parse_mode", parseMode)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
