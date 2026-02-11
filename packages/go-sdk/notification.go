package basesdk

import "context"

type SendNotificationRequest struct {
	Channel   string `json:"channel"`
	To        string `json:"to"`
	Message   string `json:"message"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func (c *Client) SendNotification(ctx context.Context, channel, to, message, parseMode string) error {
	req := SendNotificationRequest{Channel: channel, To: to, Message: message, ParseMode: parseMode}
	return c.post(ctx, "/api/v1/notifications/send", req, nil)
}
