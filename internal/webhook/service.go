package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/luikyv/mock-insurer/internal/client"
	"github.com/luikyv/mock-insurer/internal/timeutil"
)

const (
	webhookInteractionIDHeader = "X-Webhook-Interaction-Id"
	consentPath                = "/open-insurance/webhook/v1/consents/%s/consents/%s"
)

type Service struct {
	clientService client.Service
	httpClient    *http.Client
}

func NewService(clientService client.Service, httpClient *http.Client) Service {
	return Service{
		clientService: clientService,
		httpClient:    httpClient,
	}
}

func (s Service) NotifyConsent(ctx context.Context, clientID, id, version string) {
	s.notify(ctx, clientID, fmt.Sprintf(consentPath, version, id))
}

func (s Service) notify(ctx context.Context, clientID, path string) {
	client, err := s.clientService.Client(ctx, clientID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get client", "error", err)
		return
	}

	if len(client.WebhookURIs) == 0 {
		slog.DebugContext(ctx, "client has no webhook uris")
		return
	}
	webhookURI := client.WebhookURIs[0]

	data, err := json.Marshal(payload{
		Data: struct {
			Timestamp timeutil.DateTime `json:"timestamp"`
		}{
			Timestamp: timeutil.DateTimeNow(),
		},
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal webhook payload", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURI+path, bytes.NewBuffer(data))
	if err != nil {
		slog.ErrorContext(ctx, "failed to create request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(webhookInteractionIDHeader, uuid.NewString())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "failed to notify client", "error", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		slog.DebugContext(ctx, "failed to notify client", "status", resp.StatusCode)
		return
	}

	slog.InfoContext(ctx, "client was notified", "status", resp.StatusCode)
}

type payload struct {
	Data struct {
		Timestamp timeutil.DateTime `json:"timestamp"`
	} `json:"data"`
}
