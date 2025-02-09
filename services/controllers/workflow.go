package controllers

import (
	"encoding/json"
	"context"
	"fmt"
	"net/http"
	"log"
)



// ---------------- Workflow Management ----------------

// Create a Workflow
func CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	type WorkflowPayload struct {
		UserID string        `json:"user_id"`
		Tasks  []interface{} `json:"tasks"`
	}

	var payload WorkflowPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	topic := fmt.Sprintf("workflow_%s", payload.UserID)
	for _, task := range payload.Tasks {
		if err := ProduceToKafka(topic, "task", task); err != nil {
			http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Workflow created and tasks enqueued successfully"))
}

func storeData(collectionName string, data interface{}) error {
    collection := mongoClient.Database("workflow_data").Collection(collectionName)
    _, err := collection.InsertOne(context.Background(), data)
    if err != nil {
        log.Printf("Failed to store data in MongoDB: %v", err)
    }
    return err
}

// Task-specific handlers
func handleTransformTask(msg []byte) error {
    log.Printf("Handling transform task: %s", string(msg))
    transformedData := map[string]interface{}{"transformed": string(msg)}
    return storeData("transformed_tasks", transformedData)
}

func handleStoreTask(msg []byte) error {
    log.Printf("Handling store task: %s", string(msg))
    return storeData("final_tasks", msg)
}



// Workflow API
func CreateUserWorkflow(w http.ResponseWriter, r *http.Request) {
    type WorkflowRequest struct {
        UserID string        `json:"user_id"`
        Tasks  []interface{} `json:"tasks"`
    }

    var req WorkflowRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    topic := fmt.Sprintf("workflow_%s", req.UserID)
    if err := CreateKafkaTopic(topic); err != nil {
        http.Error(w, "Failed to create Kafka topic", http.StatusInternalServerError)
        return
    }

    for _, task := range req.Tasks {
        if err := ProduceToKafka(topic, "task", task); err != nil {
            http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
            return
        }
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Workflow created and tasks enqueued successfully"))
    startWorkflowConsumers(req.UserID)
}