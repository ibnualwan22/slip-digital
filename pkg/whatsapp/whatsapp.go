package whatsapp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type WhatsAppClient interface {
	SendImage(to, caption string, imageBytes []byte) error
}

type client struct {
	apiURL    string
	apiKey    string
	sessionID string
	http      *http.Client
}

func NewWhatsAppClient(apiURL, apiKey, sessionID string) WhatsAppClient {
	return &client{
		apiURL:    apiURL,
		apiKey:    apiKey,
		sessionID: sessionID,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *client) SendImage(to, caption string, imageBytes []byte) error {
	// Format to 628xxx format if it starts with 0
	if strings.HasPrefix(to, "0") {
		to = "62" + to[1:]
	}
	// Many wa-multi-session versions expect @s.whatsapp.net suffix to avoid parsing issues
	if !strings.Contains(to, "@") {
		to = to + "@s.whatsapp.net"
	}

	// Convert image to raw base64 (some APIs fail if Data URI scheme is included)
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	reqPayload := map[string]interface{}{
		"to":       to,
		"image":    base64Image,
		"filename": "slip-bisyaroh.jpg",
		"caption":  caption,
	}
	
	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("failed to encode JSON payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/sessions/%s/send/image", c.apiURL, c.sessionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	// Use Bearer or x-api-key depending on configuration. We'll set both to be safe.
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to WA API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WA API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Try to check if json response says success
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err == nil {
		if status, ok := res["status"].(bool); ok && !status {
			return fmt.Errorf("WA API returned error: %s", string(body))
		}
	}

	return nil
}
