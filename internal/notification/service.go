package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/models"
)

// Service handles notification operations
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new notification service
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// CreateChannel creates a new notification channel
func (s *Service) CreateChannel(name, channelType, config string, createdBy int64) (*models.TblNotificationChannel, error) {
	channel := &models.TblNotificationChannel{
		Name:      name,
		Type:      channelType,
		Config:    config,
		Enabled:   true,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := s.engine.Insert(channel)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

// GetChannel retrieves a channel by ID
func (s *Service) GetChannel(id int64) (*models.TblNotificationChannel, error) {
	var channel models.TblNotificationChannel
	has, err := s.engine.ID(id).Get(&channel)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("channel not found")
	}
	return &channel, nil
}

// ListChannels lists all channels
func (s *Service) ListChannels(page, pageSize int) ([]models.TblNotificationChannel, int64, error) {
	var channels []models.TblNotificationChannel
	session := s.engine.NewSession()
	defer session.Close()

	total, err := session.Count(new(models.TblNotificationChannel))
	if err != nil {
		return nil, 0, err
	}

	err = session.OrderBy("created_at DESC").Limit(pageSize, (page-1)*pageSize).Find(&channels)
	if err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// UpdateChannel updates a channel
func (s *Service) UpdateChannel(id int64, name, config string, enabled *bool) error {
	channel := &models.TblNotificationChannel{
		UpdatedAt: time.Now(),
	}
	if name != "" {
		channel.Name = name
	}
	if config != "" {
		channel.Config = config
	}
	if enabled != nil {
		channel.Enabled = *enabled
	}

	_, err := s.engine.ID(id).Cols("name", "config", "enabled", "updated_at").Update(channel)
	return err
}

// DeleteChannel deletes a channel
func (s *Service) DeleteChannel(id int64) error {
	_, err := s.engine.ID(id).Delete(new(models.TblNotificationChannel))
	return err
}

// SendNotification sends a notification to a channel
func (s *Service) SendNotification(channelId int64, notificationType, message, payload string) error {
	channel, err := s.GetChannel(channelId)
	if err != nil {
		return err
	}

	if !channel.Enabled {
		return errors.New("channel is disabled")
	}

	var sendErr error
	switch channel.Type {
	case "webhook":
		sendErr = s.sendWebhook(channel.Config, message, payload)
	case "wechat":
		sendErr = s.sendWechat(channel.Config, message, payload)
	case "feishu":
		sendErr = s.sendFeishu(channel.Config, message, payload)
	case "dingtalk":
		sendErr = s.sendDingtalk(channel.Config, message, payload)
	default:
		sendErr = fmt.Errorf("unsupported channel type: %s", channel.Type)
	}

	// Log the notification attempt
	log := &models.TblNotificationLog{
		ChannelId: channelId,
		Channel:   channel.Name,
		Type:      notificationType,
		Status:    "success",
		Message:   message,
		Payload:   payload,
		CreatedAt: time.Now(),
	}
	if sendErr != nil {
		log.Status = "failed"
		log.Message = fmt.Sprintf("%s: %v", message, sendErr)
	}

	s.engine.Insert(log)
	return sendErr
}

// sendWebhook sends a webhook notification
func (s *Service) sendWebhook(config, message, payload string) error {
	var webhookConfig struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(config), &webhookConfig); err != nil {
		return err
	}

	body := map[string]interface{}{
		"message": message,
		"payload": payload,
		"time":    time.Now().Unix(),
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(webhookConfig.URL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// sendWechat sends a WeChat Work notification
func (s *Service) sendWechat(config, message, payload string) error {
	var wechatConfig struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.Unmarshal([]byte(config), &wechatConfig); err != nil {
		return err
	}

	body := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", message, payload),
		},
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(wechatConfig.WebhookURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("wechat returned status %d", resp.StatusCode)
	}

	return nil
}

// sendFeishu sends a Feishu notification
func (s *Service) sendFeishu(config, message, payload string) error {
	var feishuConfig struct {
		WebhookURL string `json:"webhook_url"`
	}
	if err := json.Unmarshal([]byte(config), &feishuConfig); err != nil {
		return err
	}

	body := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("%s\n\n%s", message, payload),
		},
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(feishuConfig.WebhookURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("feishu returned status %d", resp.StatusCode)
	}

	return nil
}

// sendDingtalk sends a DingTalk notification
func (s *Service) sendDingtalk(config, message, payload string) error {
	var dingtalkConfig struct {
		WebhookURL string `json:"webhook_url"`
		Secret     string `json:"secret"`
	}
	if err := json.Unmarshal([]byte(config), &dingtalkConfig); err != nil {
		return err
	}

	body := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", message, payload),
		},
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(dingtalkConfig.WebhookURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("dingtalk returned status %d", resp.StatusCode)
	}

	return nil
}

// ListLogs lists notification logs
func (s *Service) ListLogs(page, pageSize int, channelId *int64) ([]models.TblNotificationLog, int64, error) {
	var logs []models.TblNotificationLog
	session := s.engine.NewSession()
	defer session.Close()

	if channelId != nil {
		session = session.Where("channel_id = ?", *channelId)
	}

	total, err := session.Count(new(models.TblNotificationLog))
	if err != nil {
		return nil, 0, err
	}

	err = session.OrderBy("created_at DESC").Limit(pageSize, (page-1)*pageSize).Find(&logs)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
