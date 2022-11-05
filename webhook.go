package switchbot

import (
	"context"
	"errors"
)

type WebhookService struct {
	c *Client
}

func newWebhookService(c *Client) *WebhookService {
	return &WebhookService{c: c}
}

func (c *Client) Webhook() *WebhookService {
	return c.webhookService
}

type webhookSetupRequest struct {
	Action     string `json:"action"`
	URL        string `json:"url,omitempty"`
	DeviceList string `json:"deviceList,omitempty"`
}

type webhookSetupResponse struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
	Message    string      `json:"message"`
}

func (svc *WebhookService) Setup(ctx context.Context, url, deviceList string) error {
	const path = "/v1.0/webhook/setupWebhook"

	if deviceList != "ALL" {
		return errors.New(`deviceList value is only supporting "ALL" for now`)
	}

	req := webhookSetupRequest{
		Action:     "setupWebhook",
		URL:        url,
		DeviceList: deviceList,
	}

	resp, err := svc.c.post(ctx, path, req)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}

type webhookQueryRequest struct {
	Action WebhookQueryActionType `json:"action"`
	URLs   []string               `json:"urls"`
}

type WebhookQueryActionType string

const (
	QueryURL     WebhookQueryActionType = "queryUrl"
	QueryDetails WebhookQueryActionType = "queryDetails"
)

type webhookQueryResponse struct {
}

func (svc *WebhookService) Query(ctx context.Context, action WebhookQueryActionType, url string) error {
	const path = "/v1.0/webhook/queryWebhook"

	req := webhookQueryRequest{
		Action: action,
	}

	switch action {
	case QueryDetails:
		if url == "" {
			return errors.New("URL need to be specified when the action is queryDetails")
		}

		req.URLs = []string{url}
	}

	resp, err := svc.c.post(ctx, path, req)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}

type webhookUpdateRequest struct {
	Action string        `json:"action"` // the type of actions but for now this must be updateWebhook
	Config webhookConfig `json:"config"`
}

type webhookConfig struct {
	URL    string `json:"url"`
	Enable bool   `json:"enable"`
}

func (svc *WebhookService) Update(ctx context.Context, url string, enable bool) error {
	const path = "/v1.0/webhook/queryWebhook"

	req := webhookUpdateRequest{
		Action: "updateWebhook",
		Config: webhookConfig{
			URL:    url,
			Enable: enable,
		},
	}

	resp, err := svc.c.post(ctx, path, req)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}

type webhookDeleteRequest struct {
	Action string `json:"action"` // the type of actions but for now this must be deleteWebhook
	URL    string `json:"url"`
}

func (svc *WebhookService) Delete(ctx context.Context, url string) error {
	const path = "/v1.0/webhook/deleteWebhook"

	req := webhookDeleteRequest{
		Action: "deleteWebhook",
		URL:    url,
	}

	resp, err := svc.c.del(ctx, path, req)
	if err != nil {
		return err
	}
	defer resp.Close()

	return nil
}
