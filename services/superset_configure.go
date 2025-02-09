package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type CSRFResponse struct {
	Result string `json:"result"`
}

func main() {
	baseURL := "http://localhost:8088" // Replace with your Superset URL
	username := "admin"        // Replace with your Superset username
	password := "admin"        // Replace with your Superset password

	// Step 1: Get Access Token
	apiToken, err := getAccessToken(baseURL, username, password)
	if err != nil {
		log.Fatalf("Failed to fetch access token: %v", err)
	}
	log.Printf("Access Token: %s", apiToken)

	client := getClientWithCookies()

	// Step 2: Get CSRF Token
	csrfToken, err := getCSRFToken(client, baseURL, apiToken)
	if err != nil {
		log.Fatalf("Failed to fetch CSRF token: %v", err)
	}
	log.Printf("CSRF Token: %s", csrfToken)

	// Step 3: Configure the dataset
	datasetPayload := map[string]interface{}{
		"database":    1,                   // Replace with your database ID
		"schema":      "workflow_data",           // Replace with your schema name
		"table_name":  "admin",             // Replace with your table name
		// "description": "Dataset for testing", // Optional description
	}

	err = configureSupersetDataset(client, baseURL, apiToken, csrfToken, datasetPayload)
	if err != nil {
		log.Fatalf("Failed to configure Superset dataset: %v", err)
	}

	log.Println("Dataset successfully configured.")
}

func getAccessToken(baseURL, username, password string) (string, error) {
	authEndpoint := fmt.Sprintf("%s/api/v1/security/login", baseURL)

	authPayload := map[string]string{
		"username": username,
		"password": password,
		"provider": "db", // Use "db" for database authentication
	}

	payloadBytes, err := json.Marshal(authPayload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize auth payload: %v", err)
	}

	req, err := http.NewRequest("POST", authEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch access token: %s", string(body))
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode access token response: %v", err)
	}

	return tokenResponse.AccessToken, nil
}

func getClientWithCookies() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func getCSRFToken(client *http.Client, baseURL, apiToken string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/security/csrf_token/", baseURL), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create CSRF token request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Referer", baseURL)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch CSRF token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch CSRF token: %s", string(body))
	}

	var csrfResponse CSRFResponse
	if err := json.NewDecoder(resp.Body).Decode(&csrfResponse); err != nil {
		return "", fmt.Errorf("failed to decode CSRF token response: %v", err)
	}

	return csrfResponse.Result, nil
}

func configureSupersetDataset(client *http.Client, baseURL, apiToken, csrfToken string, datasetPayload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(datasetPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/dataset", baseURL), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("X-CSRFToken", csrfToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Referer", baseURL)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to configure Superset dataset 1: %v", err)
	}
	log.Printf("Checkpoint 1: %s", resp.StatusCode)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to configure Superset dataset 2: %s", string(body))
	}
	log.Printf("Checkpoint 2")

	return nil
}
