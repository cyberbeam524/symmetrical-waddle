package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"strings"
	"io"
)



func CreateTrinoView(w http.ResponseWriter, r *http.Request) {
    log.Printf("createTrinoView starting")
    type ViewRequest struct {
        UserID     string `json:"user_id"`
        WorkflowID string `json:"workflow_id"`
        Collection string `json:"collection"`
    }

    var req ViewRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    query := fmt.Sprintf(`
        CREATE OR REPLACE VIEW %s AS
        SELECT * FROM mongodb.%s.%s WHERE workflow_id = '%s';
    `, req.UserID+"_view", "workflow_data", req.Collection, req.WorkflowID)

    trinoAPI := "http://localhost:8090/v1/statement"
    reqBody := strings.NewReader(query)

    // Create a new HTTP request
    httpReq, err := http.NewRequest("POST", trinoAPI, reqBody)
    if err != nil {
        log.Printf("Failed to create HTTP request: %v", err)
        http.Error(w, "Failed to create Trino view", http.StatusInternalServerError)
        return
    }

    // Add authentication headers
    httpReq.Header.Set("X-Trino-User", "your-username") // Replace with your Trino username
    httpReq.Header.Set("X-Trino-Catalog", "mongodb")    // Optional: Add catalog
    httpReq.Header.Set("X-Trino-Schema", "workflow_data") // Optional: Add schema
    httpReq.Header.Set("Content-Type", "application/json")

    // Optionally, use basic authentication
    // username := "your-username"
    // password := "your-password"
    // basicAuth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
    // httpReq.Header.Set("Authorization", "Basic "+basicAuth)

    // Make the HTTP request
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        log.Printf("Failed to create Trino view: %v", err)
        http.Error(w, "Failed to create Trino view", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Handle Trino API response
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        log.Printf("Trino API Error: %s", string(body))
        http.Error(w, "Failed to create Trino view", http.StatusInternalServerError)
        return
    }


    // Verify view creation
    verificationQuery := fmt.Sprintf(`
    SELECT table_name
    FROM information_schema.views
    WHERE table_schema = 'workflow_data'
    AND table_name = '%s';
    `, req.UserID+"_view")

    // Perform API request to check if the view exists
    verifyReq, err := http.NewRequest("POST", trinoAPI, strings.NewReader(verificationQuery))
    // (Add headers as in the previous function)
    // Add authentication headers
    verifyReq.Header.Set("X-Trino-User", "your-username") // Replace with your Trino username
    verifyReq.Header.Set("X-Trino-Catalog", "mongodb")    // Optional: Add catalog
    verifyReq.Header.Set("X-Trino-Schema", "workflow_data") // Optional: Add schema
    verifyReq.Header.Set("Content-Type", "application/json")

    resp, err = client.Do(verifyReq)
    body, _ := io.ReadAll(resp.Body)
    log.Printf("Trino Verification: %s", string(body))
    if err != nil {
        log.Printf("Verification Failed to create Trino view: %v", err)
        http.Error(w, "Verification Failed to create Trino view", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()
    


    w.WriteHeader(http.StatusOK)
    w.Write([]byte(req.UserID+"_view " + "Trino view created successfully"))
}



func CreateTrinoView2(w http.ResponseWriter, r *http.Request) {
	type ViewRequest struct {
		UserID     string `json:"user_id"`
		WorkflowID string `json:"workflow_id"`
		Collection string `json:"collection"`
	}

	var req ViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf(`
		CREATE OR REPLACE VIEW %s AS
		SELECT * FROM mongodb.%s.%s WHERE workflow_id = '%s';
	`, req.UserID+"_view", "workflow_data", req.Collection, req.WorkflowID)

	trinoAPI := "http://localhost:8090/v1/statement"
	resp, err := http.Post(trinoAPI, "application/json", strings.NewReader(query))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to create Trino view: %v", err)
		http.Error(w, "Failed to create Trino view", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trino view created successfully"))
}