package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"order/internal/domain/entities"
	"time"
)

type PaymentClient struct {
	baseURL string
}

func NewPaymentClient(url string) *PaymentClient {
	return &PaymentClient{baseURL: url}
}

func (c *PaymentClient) Authorize(orderID string, amount int64) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	body, _ := json.Marshal(map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return entities.StatusFailed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return entities.StatusFailed, nil
	}

	var result struct {
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return entities.StatusFailed, err
	}

	if result.Status == "Authorized" {
		return entities.StatusPaid, nil
	}

	return entities.StatusFailed, nil
}
