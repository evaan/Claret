package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ReadBody rejects oversized upstream responses rather than parsing truncated data.
func ReadBody(body io.Reader) ([]byte, error) {
	const limit = 8 << 20
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("upstream response exceeds 8 MiB")
	}
	return data, nil
}

func SendWebhook(webhookURL string, payload any) error {
	if webhookURL == "" {
		return nil
	}
	target, err := url.Parse(webhookURL)
	if err != nil || target.Scheme != "https" || target.Host == "" {
		return fmt.Errorf("webhook must be a valid HTTPS URL")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		// HTTP errors contain the URL, including Discord's secret token.
		return fmt.Errorf("webhook request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
