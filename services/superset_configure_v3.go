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

type DatasetPayload struct {
	Database    int    `json:"database"`
	Schema      string `json:"schema"`
	TableName   string `json:"table_name"`
	Owners      []int  `json:"owners"`
	// Extra       map[string]interface{} `json:"extra"` // To enforce RLS
}

func createRLSRule(client *http.Client, baseURL, apiToken, csrfToken string, datasetID int, userID int) error {
	rlsEndpoint := fmt.Sprintf("%s/api/v1/row_level_security", baseURL)

	rlsPayload := map[string]interface{}{
		"dataset_id": datasetID,
		"filter_type": "Regular",
		"clause":      fmt.Sprintf("owner_id = %d", userID),
		"roles":       []int{}, // Specify roles allowed to bypass this rule (optional)
	}

	payloadBytes, err := json.Marshal(rlsPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize RLS payload: %v", err)
	}

	req, err := http.NewRequest("POST", rlsEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create RLS rule request: %v", err)
	}

	req.Header.Set("X-CSRFToken", csrfToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create RLS rule: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create RLS rule: %s", string(body))
	}

	return nil
}


func main() {
	baseURL := "http://localhost:8088" // Replace with your Superset URL
	username := "admin"               // Replace with your Superset username
	password := "admin"               // Replace with your Superset password
	userID := 3                       // Replace with the user's ID

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
	// rlsFilter := fmt.Sprintf("{\"filter_type\":\"Row\",\"clause\":\"owner_id = %d\"}", userID)
	datasetPayload := DatasetPayload{
		Database:    1,               // Replace with your database ID
		Schema:      "workflow_data", // Replace with your schema name
		TableName:   "user_data_table_1", // Replace with the user's table name
		Owners:      []int{userID},   // Assign dataset ownership to the user
		// Extra: map[string]interface{}{
		// 	"metadata_params": map[string]string{"rls": rlsFilter},
		// }, // Add RLS enforcement
	}

	err = configureSupersetDataset(client, baseURL, apiToken, csrfToken, datasetPayload)
	if err != nil {
		log.Fatalf("Failed to configure Superset dataset: %v", err)
	}

	log.Println("Dataset successfully configured")

	// Step 4: Create RLS rule for the dataset
	datasetID := 1 // Replace with the actual dataset ID after creation
	err = createRLSRule(client, baseURL, apiToken, csrfToken, datasetID, userID)
	if err != nil {
		log.Fatalf("Failed to create RLS rule: %v", err)
	}
	log.Println("RLS rule successfully created for the user.")

	
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

func configureSupersetDataset(client *http.Client, baseURL, apiToken, csrfToken string, datasetPayload DatasetPayload) error {
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
		return fmt.Errorf("failed to configure Superset dataset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to configure Superset dataset: %s", string(body))
	}

	return nil
}
