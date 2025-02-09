package controllers

import (

	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"io"
	"net/http/cookiejar"
	"bytes"
)



// superset functions:

func createSupersetDataset(datasetName, tableName, databaseID, userToken string) error {
	apiURL := "http://localhost:8088/api/v1/dataset/"
	payload := map[string]interface{}{
		"database": databaseID,
		"table_name": tableName,
		"schema":    "public",
		"sql":       nil,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create dataset in Superset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Superset dataset creation failed: %s", string(body))
	}

	log.Println("Dataset created successfully")
	return nil
}

func createSupersetDashboard(dashboardName, userToken string) error {
	apiURL := "http://localhost:8088/api/v1/dashboard/"
	payload := map[string]interface{}{
		"dashboard_title": dashboardName,
		"position_json":   nil, // Optional layout
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create dashboard in Superset: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Superset dashboard creation failed: %s", string(body))
	}

	log.Println("Dashboard created successfully")
	return nil
}







type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type CSRFResponse struct {
	Result string `json:"result"`
}

// Fetch API token by logging into Superset
func getAPIToken(supersetAPIBaseURL, username, password string) (string, error) {
	loginEndpoint := fmt.Sprintf("%s/api/v1/security/login", supersetAPIBaseURL)
	payload := map[string]string{
		"username": username,
		"password": password,
		"provider": "db", // For database authentication
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to serialize login payload: %v", err)
	}

	req, err := http.NewRequest("POST", loginEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to login: %s", string(body))
	}

	var loginResponse LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		return "", fmt.Errorf("failed to decode login response: %v", err)
	}

	return loginResponse.AccessToken, nil
}

// Fetch CSRF token using the API token
func getCSRFToken(supersetAPIBaseURL, apiToken string) (string, error) {
	csrfEndpoint := fmt.Sprintf("%s/api/v1/security/csrf_token/", supersetAPIBaseURL)
	req, err := http.NewRequest("GET", csrfEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create CSRF token request: %v", err)
	}

	// Add API token in Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Referer", supersetAPIBaseURL) // Required by Superset

	resp, err := http.DefaultClient.Do(req)
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




func getClientWithCookies() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func configureSupersetDataset(baseURL, apiToken, csrfToken string, datasetPayload map[string]interface{}) error {
	// Create JSON payload
	payloadBytes, err := json.Marshal(datasetPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/dataset", baseURL), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	req.Header.Set("X-CSRFToken", csrfToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	req.Header.Set("Referer", baseURL)

	// Use a client with cookies enabled
	client := getClientWithCookies()
	resp, err := client.Do(req)
    log.Println("Failed to configure Superset dataset 222")
	if err != nil {
		return fmt.Errorf("failed to configure Superset dataset: %v", err)
	}
	defer resp.Body.Close()


    log.Println("Failed to configure Superset dataset 333")
	// Check response status
	if resp.StatusCode != http.StatusCreated {
        log.Println("Failed to configure Superset dataset 444")
		body, _ := io.ReadAll(resp.Body)
        log.Println("failed to configure Superset dataset: %s", string(body))
		return fmt.Errorf("failed to configure Superset dataset: %s", string(body))
	}

    log.Println("Failed to configure Superset dataset 555")

	return nil
}

func HandleConfigureSupersetDataset(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse the request body
    var reqBody struct {
        BaseURL       string                 `json:"base_url"`
        ApiToken      string                 `json:"api_token"`
        DatasetPayload map[string]interface{} `json:"dataset_payload"`
    }

    if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }



    // Step 1: Get API token
    supersetAPIBaseURL := "http://localhost:8088"
	username := "admin"
	password := "admin"

	apiToken, err := getAPIToken(supersetAPIBaseURL, username, password)
	if err != nil {
		fmt.Printf("Error fetching API token: %v\n", err)
		return
	}
	fmt.Printf("API Token: %s\n", apiToken)

        // Get CSRF token
        csrfToken, err := getCSRFToken(reqBody.BaseURL, apiToken)
        if err != nil {
            log.Println("Failed to fetch CSRF token")
            http.Error(w, fmt.Sprintf("Failed to fetch CSRF token: %v", err), http.StatusInternalServerError)
            return
        }

        log.Println("csrfToken:%s", csrfToken)

    // Call the function to configure the dataset
    err = configureSupersetDataset(reqBody.BaseURL, apiToken, csrfToken, reqBody.DatasetPayload)
    if err != nil {
        log.Println("Failed to configure Superset dataset")
        http.Error(w, fmt.Sprintf("Failed to configure Superset dataset: %v", err), http.StatusInternalServerError)
        return
    }

    // Respond with success
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Superset dataset configured successfully"))
}



func ConfigureSupersetDatasetlatest(w http.ResponseWriter, r *http.Request) {
	type DatasetRequest struct {
		CollectionName string `json:"collection_name"` // Specify the collection name
		UserID         string `json:"user_id"`        // User ID to associate with the dataset
	}

	// Parse the incoming request
	var reqBody DatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Prepare the API URL and the payload
    supersetAPIBaseURL2 := "http://localhost:8088/api/v1"
    supersetAPI := fmt.Sprintf("%s/dataset", supersetAPIBaseURL2)

    // Step 1: Fetch CSRF Token
    // csrfToken, err := getCSRFToken(supersetAPIBaseURL)
    // if err != nil {
    //     http.Error(w, fmt.Sprintf("Failed to fetch CSRF token: %v", err), http.StatusInternalServerError)
    //     return
    // }


    	// Superset configuration
	supersetAPIBaseURL := "http://localhost:8088"
	username := "admin"
	password := "admin"

	// Step 1: Get API token
	apiToken, err := getAPIToken(supersetAPIBaseURL, username, password)
	if err != nil {
		fmt.Printf("Error fetching API token: %v\n", err)
		return
	}
	fmt.Printf("API Token: %s\n", apiToken)

	// Step 2: Get CSRF token
	csrfToken, err := getCSRFToken(supersetAPIBaseURL, apiToken)
	if err != nil {
		fmt.Printf("Error fetching CSRF token: %v\n", err)
		return
	}
	fmt.Printf("CSRF Token: %s\n", csrfToken)

	payload := map[string]interface{}{
		"database":    1, // Replace with the correct Trino database ID in Superset
		"schema":      "mongodb.workflow_data", // Adjust the schema name to your actual Trino schema
		"table_name":  reqBody.CollectionName, // Use the provided collection name
		"description": fmt.Sprintf("Dataset for user %s and collection %s", reqBody.UserID, reqBody.CollectionName),
	}

	// Serialize the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to serialize payload", http.StatusInternalServerError)
		return
	}

	// Create the HTTP request for Superset API
	httpReq, err := http.NewRequest("POST", supersetAPI, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Failed to create HTTP request", http.StatusInternalServerError)
		return
	}

	// Set the necessary headers, including Superset API token
    httpReq.Header.Set("Referer", "http://localhost:8088") // Replace with your Superset host
    httpReq.Header.Set("X-CSRFToken", csrfToken)
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	// Execute the HTTP request
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		http.Error(w, "Failed to configure Superset dataset", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status from Superset API
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Failed to configure Superset dataset: %s", string(body)), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Superset dataset for collection %s configured successfully", reqBody.CollectionName)))
}


func ConfigureSupersetDataset2(w http.ResponseWriter, r *http.Request) {
	type DatasetRequest struct {
		ViewName string `json:"view_name"`
		UserID   string `json:"user_id"`
	}

	var reqBody DatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	supersetAPI := "http://localhost:8088/api/v1/dataset"
	payload := map[string]interface{}{
		"database":    1, // Replace with your Trino database ID in Superset
		"schema":      "mongodb",
		"table_name":  reqBody.ViewName,
		"description": fmt.Sprintf("Dataset for user %s", reqBody.UserID),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to serialize payload", http.StatusInternalServerError)
		return
	}

	httpReq, err := http.NewRequest("POST", supersetAPI, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Failed to create HTTP request", http.StatusInternalServerError)
		return
	}

	httpReq.Header.Set("Authorization", "Bearer your_superset_api_token") // Use your Superset API token
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil || resp.StatusCode != http.StatusCreated {
		http.Error(w, "Failed to configure Superset dataset", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Superset dataset configured successfully"))
}




