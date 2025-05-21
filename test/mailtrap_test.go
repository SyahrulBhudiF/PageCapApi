package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"
)

const (
	mailtrapAccountID = "1956740"
	mailtrapInboxID   = "2950828"
	mailtrapAPIToken  = "b5d22f00f2c7f6d90da907d59415ab5f"
)

func fetchMailtrapContent(t *testing.T, contentPath string) (string, error) {
	if contentPath == "" {
		return "", fmt.Errorf("content path is empty")
	}
	url := "https://mailtrap.io" + contentPath

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for content: %w", err)
	}

	req.Header.Set("Api-Token", mailtrapAPIToken)
	req.Header.Set("Accept", "text/plain, text/html, */*")

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch content: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read content body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code %d: %s", resp.StatusCode, TruncateString(string(bodyBytes), 200))
	}

	return string(bodyBytes), nil
}

func getLatestOTPFromMailtrap(t *testing.T) (string, error) {
	listMessagesURL := fmt.Sprintf(
		"https://mailtrap.io/api/accounts/%s/inboxes/%s/messages",
		mailtrapAccountID,
		mailtrapInboxID,
	)

	req, err := http.NewRequest("GET", listMessagesURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create message list request: %w", err)
	}

	req.Header.Set("Api-Token", mailtrapAPIToken)
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch message list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	type MailtrapMessage struct {
		ID           int    `json:"id"`
		Subject      string `json:"subject"`
		HTMLPath     string `json:"html_path"`
		TxtPath      string `json:"txt_path"`
		HTMLBodySize int    `json:"html_body_size"`
		TextBodySize int    `json:"text_body_size"`
		CreatedAt    string `json:"created_at"`
	}
	var messages []MailtrapMessage

	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return "", fmt.Errorf("failed to parse message list: %w", err)
	}

	if len(messages) == 0 {
		return "", fmt.Errorf("no messages found")
	}

	latestMessage := messages[0]
	t.Logf("Latest message: ID %d, Subject: %s", latestMessage.ID, latestMessage.Subject)

	re := regexp.MustCompile(`\b\d{6}\b`)
	var otp string

	if latestMessage.HTMLBodySize > 0 && latestMessage.HTMLPath != "" {
		htmlContent, err := fetchMailtrapContent(t, latestMessage.HTMLPath)
		if err == nil {
			matches := re.FindStringSubmatch(htmlContent)
			if len(matches) > 0 {
				otp = matches[0]
				t.Logf("OTP found in HTML: %s", otp)
				return otp, nil
			}
		}
	}

	if latestMessage.TextBodySize > 0 && latestMessage.TxtPath != "" {
		textContent, err := fetchMailtrapContent(t, latestMessage.TxtPath)
		if err != nil {
			return "", fmt.Errorf("failed to fetch text content: %w", err)
		}
		matches := re.FindStringSubmatch(textContent)
		if len(matches) > 0 {
			otp = matches[0]
			t.Logf("OTP found in text: %s", otp)
			return otp, nil
		}
	}

	return "", fmt.Errorf("OTP not found")
}
