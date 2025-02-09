package controllers

import (
	"bytes"
	// "context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "os"
	// "strings"
	// "github.com/gorilla/mux"
	// "github.com/playwright-community/playwright-go"
	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
	// "github.com/segmentio/kafka-go"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	// "services/scraper/controllers"
	// "services/scraper/config"
	// "github.com/joho/godotenv"
    // "github.com/gorilla/handlers"
)



// ---------------- Spark Integration ----------------

// Submit a Spark Job
func SubmitSparkJob(w http.ResponseWriter, r *http.Request) {
	type SparkJobPayload struct {
		JobName   string   `json:"job_name"`
		Arguments []string `json:"arguments"`
	}

	var payload SparkJobPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	sparkAPI := "http://spark-master:8083/v1/submissions/create"
	reqBody, _ := json.Marshal(map[string]interface{}{
		"action": "CreateSubmissionRequest",
		"appArgs": payload.Arguments,
		"appResource": fmt.Sprintf("local:/opt/spark-jobs/%s", payload.JobName),
	})

	resp, err := http.Post(sparkAPI, "application/json", bytes.NewBuffer(reqBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to submit Spark job: %v", err)
		http.Error(w, "Failed to submit Spark job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Spark job submitted successfully"))
}

func TriggerSparkJob(w http.ResponseWriter, r *http.Request) {
    jobData := map[string]string{
        "job_type": "spark_transformation",
    }
    jobDataJSON, _ := json.Marshal(jobData)
    err := ProduceToKafka("spark_jobs", "spark_task", jobDataJSON)
    if err != nil {
        http.Error(w, "Failed to trigger Spark job", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Spark job triggered successfully"))
}