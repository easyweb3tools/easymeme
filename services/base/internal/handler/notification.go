package handler

import (
	"fmt"
	"net/http"

	"easyweb3/base/internal/model"
	"easyweb3/base/internal/repository"
	"easyweb3/base/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	telegram *service.TelegramClient
	repo     *repository.Repository
}

func NewNotificationHandler(telegram *service.TelegramClient, repo *repository.Repository) *NotificationHandler {
	return &NotificationHandler{telegram: telegram, repo: repo}
}

type SendNotificationRequest struct {
	Channel   string `json:"channel" binding:"required"`
	To        string `json:"to" binding:"required"`
	Message   string `json:"message" binding:"required"`
	ParseMode string `json:"parse_mode"`
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceID, _ := c.Get("service_id")
	serviceIDStr, _ := serviceID.(string)

	sent := false
	sendErr := ""
	switch req.Channel {
	case "telegram":
		if err := h.telegram.SendMessage(c.Request.Context(), req.To, req.Message, req.ParseMode); err != nil {
			sendErr = err.Error()
		} else {
			sent = true
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported channel: %s", req.Channel)})
		return
	}

	status := "sent"
	if !sent {
		status = "failed"
	}
	_ = h.repo.CreateNotificationLog(c.Request.Context(), &model.NotificationLog{
		ServiceID: serviceIDStr,
		Channel:   req.Channel,
		Recipient: req.To,
		Message:   req.Message,
		Status:    status,
	})

	if !sent {
		c.JSON(http.StatusBadGateway, gin.H{"error": sendErr})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "sent"}})
}
