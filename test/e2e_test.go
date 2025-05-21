package test

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

func TestE2E(t *testing.T) {
	testEmail := fmt.Sprintf("testuser_@example.com")
	testPassword := "Password123!"
	testName := "Test User"

	// --- 1. Register a new user ---
	t.Run("Register User", func(t *testing.T) {
		registerPayload := map[string]string{
			"email":    testEmail,
			"password": testPassword,
			"name":     testName,
			"confirm":  testPassword,
		}
		resp, body, err := makeRequest("POST", baseURL+"/auth/register", registerPayload, nil)
		if err != nil {
			t.Fatalf("Registration request failed: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d. Response: %s", http.StatusCreated, resp.StatusCode, string(body))
		} else {
			t.Log("User registered successfully")
		}
	})

	// --- 2. Send OTP ---
	t.Run("Send OTP", func(t *testing.T) {
		sendOTPPayload := map[string]string{"email": testEmail}
		resp, body, err := makeRequest("POST", baseURL+"/auth/send-otp", sendOTPPayload, nil)
		if err != nil {
			t.Fatalf("Send OTP request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		} else {
			t.Log("OTP sent successfully")
		}
	})

	// --- 3. Verify Email with OTP ---
	t.Run("VerifyEmail", func(t *testing.T) {
		var otp string
		var err error
		for i := 0; i < 10; i++ {
			otp, err = getLatestOTPFromMailtrap(t)
			if err == nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
		if err != nil {
			t.Fatalf("Failed to get OTP: %v", err)
		}

		payload := map[string]string{
			"email": testEmail,
			"otp":   otp,
		}
		resp, body, err := makeRequest("POST", baseURL+"/auth/verify-email", payload, nil)
		if err != nil {
			t.Fatalf("Verification failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, body)
		}
	})

	// --- 4. Login with verified user ---
	var accessToken, refreshToken string
	t.Run("Login User", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}
		resp, body, err := makeRequest("POST", baseURL+"/auth/login", loginPayload, nil)
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		var loginResponse map[string]interface{}
		if err := json.Unmarshal(body, &loginResponse); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		dataMap, ok := loginResponse["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Missing or invalid data field in response")
		}

		var ok1, ok2 bool
		accessToken, ok1 = dataMap["access_token"].(string) // ✅ BUKAN := tapi =
		refreshToken, ok2 = dataMap["refresh_token"].(string)

		if !ok1 || !ok2 {
			t.Fatalf("Missing or invalid access_token or refresh_token")
		}

		if accessToken == "" || refreshToken == "" {
			t.Fatalf("Access token or refresh token is empty")
		} else {
			t.Logf("Login successful. Access Token: %s... Refresh Token: %s...", accessToken[:10], refreshToken[:10])
		}
	})

	// --- 5. Use Access Token to Generate API Key ---
	var apiKey string
	t.Run("Generate API Key", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping Generate API Key test as access token was not obtained.")
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		resp, body, err := makeRequest("GET", baseURL+"/auth/api-key", nil, headers)
		if err != nil {
			t.Fatalf("Generate API Key request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		type APIKeyResponse struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Data    struct {
				APIKey string `json:"api_key"`
			} `json:"data"`
		}

		var respData APIKeyResponse
		if err := json.Unmarshal(body, &respData); err != nil {
			t.Fatalf("Failed to decode API key response: %v", err)
		}
		apiKey = fmt.Sprintf("%v", respData.Data.APIKey)

		if apiKey == "" {
			t.Errorf("API key response missing api_key")
		} else {
			t.Logf("API Key generated successfully: %s...", apiKey[:10])
		}
	})

	// --- 6. Test Page Capture ---
	t.Run("Page Capture", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping Page Capture test as access token was not obtained.")
		}

		pageCapturePayload := map[string]interface{}{
			"url":      "https://example.com",
			"fullPage": true,
		}

		requestURL := baseURL + "/page-capture/" + apiKey
		t.Logf("Attempting Page Capture request to URL: %s with payload: %+v", requestURL, pageCapturePayload)

		resp, body, err := makeRequest("POST", requestURL, pageCapturePayload, nil)
		if err != nil {
			t.Fatalf("Page Capture request failed: %v", err)
		}

		t.Logf("Page Capture request completed with status code: %d", resp.StatusCode)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		} else {
			t.Log("Page captured successfully")
			if resp.Header.Get("Content-Type") != "image/png" {
				t.Errorf("Expected Content-Type 'image/png', got '%s'", resp.Header.Get("Content-Type"))
			}
		}
	})

	// --- 7. Refresh Token ---
	t.Run("Refresh Token", func(t *testing.T) {
		if refreshToken == "" {
			t.Skip("Skipping Refresh Token test as refresh token was not obtained.")
		}

		refreshPayload := map[string]string{
			"refresh_token": refreshToken,
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		resp, body, err := makeRequest("POST", baseURL+"/auth/refresh", refreshPayload, headers)
		if err != nil {
			t.Fatalf("Refresh Token request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		type RefreshResponse struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Data    struct {
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}

		fmt.Println(string(body))

		var refreshResponse RefreshResponse
		if err := json.Unmarshal(body, &refreshResponse); err != nil {
			t.Fatalf("Failed to decode refresh response: %v", err)
		}

		accessToken = refreshResponse.Data.AccessToken

		if accessToken == "" {
			t.Fatalf("New access token is empty after refresh")
		} else {
			t.Logf("Token refreshed successfully. New Access Token: %s... ", accessToken[:10])
		}
	})

	// --- 8. Get Page Capture History (using new access token) ---
	t.Run("Get Page Capture History", func(t *testing.T) {
		time.Sleep(3 * time.Second)

		if accessToken == "" {
			t.Skip("Skipping Get Page Capture History test as access token was not obtained after refresh.")
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}

		resp, body, err := makeRequest("GET", baseURL+"/page-capture", nil, headers)
		if err != nil {
			t.Fatalf("Get Page Capture History request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		}

		type PageCapture struct {
			UserID       uuid.UUID `json:"user_id"`
			URL          string    `json:"url"`
			ImagePath    string    `json:"image_path"`
			PublicId     string    `json:"public_id"`
			Width        *int      `json:"width,omitempty"`
			Height       *int      `json:"height,omitempty"`
			FullPage     bool      `json:"full_page"`
			DelaySeconds int       `json:"delay_seconds"`
			IsMobile     bool      `json:"is_mobile"`
		}

		type PagesCaptureData struct {
			Data       []PageCapture `json:"history"`
			Total      int64         `json:"total"`
			Page       int           `json:"page"`
			Limit      int           `json:"limit"`
			TotalPages int           `json:"total_pages"`
		}

		type PagesCaptureResponse struct {
			Data PagesCaptureData `json:"data"`
		}

		var historyResponse PagesCaptureResponse
		if err := json.Unmarshal(body, &historyResponse); err != nil {
			t.Fatalf("Failed to decode history response: %v", err)
		}

		data := historyResponse.Data.Data
		if len(data) == 0 {
			t.Log("No page capture history found (this might be expected if cleanup occurs)")
		} else {
			t.Logf("Successfully retrieved page capture history with %d items", len(data))
		}
	})

	// --- 9. Logout ---
	t.Run("Logout User", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping Logout test as access token was not obtained.")
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		refreshPayload := map[string]string{
			"refresh_token": refreshToken,
		}

		resp, body, err := makeRequest("DELETE", baseURL+"/auth/logout", refreshPayload, headers)
		if err != nil {
			t.Fatalf("Logout request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status code %d, got %d. Response: %s", http.StatusOK, resp.StatusCode, string(body))
		} else {
			t.Log("User logged out successfully")
		}
	})

	// --- 10. Attempt to Generate API Key after Logout (Expected Unauthorized) ---
	t.Run("Generate API Key After Logout (Unauthorized)", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping Generate API Key After Logout test as access token was not obtained.")
		}

		headers := map[string]string{
			"Authorization": "Bearer " + accessToken,
		}
		resp, body, err := makeRequest("GET", baseURL+"/auth/api-key", nil, headers)
		if err != nil {
			t.Logf("Generate API Key request after logout failed as expected: %v", err)
		}

		// Expecting Unauthorized status code
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status code %d (Unauthorized), got %d. Response: %s", http.StatusUnauthorized, resp.StatusCode, string(body))
		} else {
			t.Log("Attempt to generate API Key after logout correctly returned Unauthorized")
		}
	})
}
