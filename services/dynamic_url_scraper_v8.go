package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"text/template"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variables
var (
    kafkaWriter       *kafka.Writer
    mongoClient       *mongo.Client
    taskProcessedCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "task_processed_total",
            Help: "Total number of tasks processed by type",
        },
        []string{"task_type"},
    )
)

// Initialization functions
func init() {
    initKafka()
    initMongo()
    initPrometheus()
}

func initKafka() {
    kafkaWriter = &kafka.Writer{
        Addr:     kafka.TCP("localhost:9092"),
        Balancer: &kafka.LeastBytes{},
    }
    log.Println("Kafka initialized")
}

func initMongo() {
    var err error
    mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    log.Println("MongoDB initialized")
}

func initPrometheus() {
    prometheus.MustRegister(taskProcessedCounter)
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(":9091", nil)
    log.Println("Prometheus metrics server started on :9091")
}

// Utility functions
func incrementTaskCounter(taskType string) {
    taskProcessedCounter.WithLabelValues(taskType).Inc()
}

func createKafkaTopic(topic string) error {
    conn, err := kafka.Dial("tcp", "localhost:9092")
    if err != nil {
        return fmt.Errorf("Failed to connect to Kafka: %v", err)
    }
    defer conn.Close()

    err = conn.CreateTopics(kafka.TopicConfig{
        Topic:             topic,
        NumPartitions:     3,
        ReplicationFactor: 1,
    })
    if err != nil {
        return fmt.Errorf("Failed to create Kafka topic: %v", err)
    }
    log.Printf("Kafka topic '%s' created successfully", topic)
    return nil
}


// ---------------- Kafka Integration ----------------

func produceToKafka(topic string, taskType string, message interface{}) error {
	msg, err := json.Marshal(map[string]interface{}{
		"task_type": taskType,
		"payload":   message,
	})
	if err != nil {
		return fmt.Errorf("Failed to serialize message: %v", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
	})
	err = writer.WriteMessages(context.Background(), kafka.Message{Value: msg})
	if err != nil {
		return fmt.Errorf("Failed to write message to Kafka: %v", err)
	}

	log.Printf("Message for task '%s' written to topic '%s'", taskType, topic)
	return nil
}

// ---------------- Workflow Management ----------------

// Create a Workflow
func createWorkflow(w http.ResponseWriter, r *http.Request) {
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
		if err := produceToKafka(topic, "task", task); err != nil {
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

// Kafka consumer
func startConsumer(topic, taskType string, handler func(msg []byte) error) {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: []string{"localhost:9092"},
        Topic:   topic,
        GroupID: fmt.Sprintf("%s_consumer", taskType),
    })

    go func() {
        for {
            msg, err := reader.ReadMessage(context.Background())
            if err != nil {
                log.Printf("Error reading message: %v", err)
                continue
            }

            var task map[string]interface{}
            if err := json.Unmarshal(msg.Value, &task); err != nil {
                log.Printf("Failed to parse message: %v", err)
                continue
            }

            if task["task_type"] == taskType {
                if err := handler(msg.Value); err != nil {
                    log.Printf("Task handling failed: %v", err)
                } else {
                    incrementTaskCounter(taskType)
                }
            }
        }
    }()
}

// Workflow API
func createUserWorkflow(w http.ResponseWriter, r *http.Request) {
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
    if err := createKafkaTopic(topic); err != nil {
        http.Error(w, "Failed to create Kafka topic", http.StatusInternalServerError)
        return
    }

    for _, task := range req.Tasks {
        if err := produceToKafka(topic, "task", task); err != nil {
            http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
            return
        }
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Workflow created and tasks enqueued successfully"))
    startWorkflowConsumers(req.UserID)
}

func startWorkflowConsumers(userID string) {
    topic := fmt.Sprintf("workflow_%s", userID)
    startConsumer(topic, "transform", handleTransformTask)
    startConsumer(topic, "store", handleStoreTask)
}

func triggerSparkJob(w http.ResponseWriter, r *http.Request) {
    jobData := map[string]string{
        "job_type": "spark_transformation",
    }
    jobDataJSON, _ := json.Marshal(jobData)
    err := produceToKafka("spark_jobs", "spark_task", jobDataJSON)
    if err != nil {
        http.Error(w, "Failed to trigger Spark job", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Spark job triggered successfully"))
}


// ---------------- Airflow Integration ----------------

// Trigger an Airflow DAG
func triggerAirflowDag(w http.ResponseWriter, r *http.Request) {
	type DagPayload struct {
		DagID  string                 `json:"dag_id"`
		Params map[string]interface{} `json:"params"`
	}

	var payload DagPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	airflowAPI := fmt.Sprintf("http://localhost:8081/api/v1/dags/%s/dagRuns", payload.DagID)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"conf": payload.Params,
	})

	resp, err := http.Post(airflowAPI, "application/json", bytes.NewBuffer(reqBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to trigger Airflow DAG: %v", err)
		http.Error(w, "Failed to trigger Airflow DAG", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Airflow DAG triggered successfully"))
}

// ---------------- Spark Integration ----------------

// Submit a Spark Job
func submitSparkJob(w http.ResponseWriter, r *http.Request) {
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



var taskDurationHistogram = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "task_duration_seconds",
        Help: "Duration of each task type in seconds",
    },
    []string{"task_type"},
)

func recordTaskDuration(taskType string, duration float64) {
    taskDurationHistogram.WithLabelValues(taskType).Observe(duration)
}


type Task struct {
	TaskID     string   `json:"task_id"`
	TaskType   string   `json:"task_type"`
	DependsOn  []string `json:"depends_on"`
}

type Workflow struct {
	UserID     string `json:"user_id"`
	WorkflowID string `json:"workflow_id"`
    Schedule  string // User-defined schedule
	Tasks      []Task `json:"tasks"`
}

// Endpoint to submit a workflow
func submitWorkflow(w http.ResponseWriter, r *http.Request) {
	var workflow Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

    // Validate the workflow payload
	if workflow.UserID == "" || workflow.WorkflowID == "" {
		http.Error(w, "Missing required fields: user_id or workflow_id", http.StatusBadRequest)
		return
	}

	if workflow.Tasks == nil || len(workflow.Tasks) == 0 {
		http.Error(w, "Workflow must contain at least one task", http.StatusBadRequest)
		return
	}



	// Save workflow configuration to file
	filePath := filepath.Join("workflows", fmt.Sprintf("%s_%s.json", workflow.UserID, workflow.WorkflowID))
	if err := os.MkdirAll("workflows", os.ModePerm); err != nil {
		http.Error(w, "Failed to create directory for workflows", http.StatusInternalServerError)
		return
	}
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(workflow); err != nil {
		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
		return
	}

	// Trigger DAG creation in Airflow
	if err := triggerAirflowDagCreation(workflow); err != nil {
		http.Error(w, "Failed to trigger Airflow DAG creation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Workflow submitted successfully"))
}

// Trigger Airflow DAG creation
func triggerAirflowDagCreation1(workflow Workflow) error {
    log.Println("starting triggerAirflowDagCreation")

	airflowURL := "http://localhost:8081/api/v1/dags"
	dagPayload := map[string]interface{}{
		"dag_id":      fmt.Sprintf("workflow_%s_%s", workflow.UserID, workflow.WorkflowID),
		"description": "User-defined workflow DAG",
	}
	payloadBytes, _ := json.Marshal(dagPayload)
    log.Println("triggerAirflowDagCreation posting to airflow")
	resp, err := http.Post(airflowURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
        log.Println("failed to trigger Airflow DAG creation")
		return fmt.Errorf("failed to trigger Airflow DAG creation: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
        log.Println("Airflow DAG creation failed with status %d", resp.StatusCode)
		return fmt.Errorf("Airflow DAG creation failed with status: %d", resp.StatusCode)
	}
	return nil
}

// Reserialize DAGs to reflect the latest changes
func triggerDAGReserialize() error {
	cmd := exec.Command("docker", "exec", "airflow", "airflow", "dags", "reserialize")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to reserialize DAGs: %v, Output: %s", err, string(output))
		return err
	}
	log.Printf("Successfully reserialized DAGs: %s", string(output))
	return nil
}

// Trigger a specific DAG
func triggerAirflowDAG(dagID string) error {
	cmd := exec.Command("docker", "exec", "airflow", "airflow", "dags", "trigger", dagID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to trigger DAG '%s': %v, Output: %s", dagID, err, string(output))
		return err
	}
	log.Printf("Successfully triggered DAG '%s': %s", dagID, string(output))
	return nil
}

// Unpause a specific DAG
func unpauseAirflowDAG(dagID string) error {
	cmd := exec.Command("docker", "exec", "airflow", "airflow", "dags", "unpause", dagID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to unpause DAG '%s': %v, Output: %s", dagID, err, string(output))
		return err
	}
	log.Printf("Successfully unpaused DAG '%s': %s", dagID, string(output))
	return nil
}



func triggerAirflowDagCreation(workflow Workflow) error {
	log.Println("starting triggerAirflowDagCreation")

	// Define the path to the Airflow `dags` folder
	dagFilePath := fmt.Sprintf("../dags/workflow_%s_%s.py", workflow.UserID, workflow.WorkflowID)

	// Prepare the DAG payload for the template
	dagData := struct {
		DagID       string
		Description string
        Schedule    string
		Tasks       []struct {
			TaskID string
		}
	}{
		DagID:       fmt.Sprintf("workflow_%s_%s", workflow.UserID, workflow.WorkflowID),
		Description: "User-defined workflow DAG",
        Schedule:    workflow.Schedule, // User-defined schedule
	}

	// Add tasks to the DAG
	for _, task := range workflow.Tasks {
		dagData.Tasks = append(dagData.Tasks, struct {
			TaskID string
		}{
			TaskID: task.TaskID,
		})
	}

	// Write the DAG file
	tmpl, err := template.New("airflowDag").Parse(airflowDagTemplate)
	if err != nil {
        log.Println("failed to parse DAG template: %v", err)
		return fmt.Errorf("failed to parse DAG template: %v", err)
	}
	file, err := os.Create(dagFilePath)
	if err != nil {
        log.Println("failed to create DAG file: %v", err)
		return fmt.Errorf("failed to create DAG file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, dagData); err != nil {
        log.Println("failed to write DAG file: %v", err)
		return fmt.Errorf("failed to write DAG file: %v", err)
	}


    // Reserialize DAGs
	if err := triggerDAGReserialize(); err != nil {
		return fmt.Errorf("Failed to reserialize DAGs: %v", err)
	}

    // Unpause the specific DAG
	dagID := fmt.Sprintf("workflow_%s_%s", workflow.UserID, workflow.WorkflowID)

	// Trigger the specific DAG
	if err := triggerAirflowDAG(dagID); err != nil {
		return fmt.Errorf("Failed to trigger DAG '%s': %v", dagID, err)
	}

    if err := unpauseAirflowDAG(dagID); err != nil {
		return fmt.Errorf("Failed to unpause DAG '%s': %v", dagID, err)
	}

    
	log.Println("DAG file created successfully:", dagFilePath)

	return nil
}


func triggerAirflowDagRun(dagID string) error {
	airflowURL := fmt.Sprintf("http://localhost:8081/api/v1/dags/%s/dagRuns", dagID)
	payload := map[string]interface{}{
		"conf": map[string]string{},
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(airflowURL, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to trigger Airflow DAG run: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Airflow DAG run failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	log.Println("Airflow DAG run triggered successfully")
	return nil
}


func checkAirflowHealth() error {
    resp, err := http.Get("http://localhost:8081/health")
    if err != nil {
        return fmt.Errorf("Airflow not reachable: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("Airflow health check failed: %d", resp.StatusCode)
    }
    return nil
}


// Template for the dynamic Airflow DAG
const airflowDagTemplate = `
from airflow import DAG
from airflow.operators.dummy import DummyOperator
from airflow.operators.python import PythonOperator
from datetime import datetime, timedelta

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

def task_function(task_id):
    print(f"Executing task: {task_id}")

with DAG(
    dag_id="{{.DagID}}",
    default_args=default_args,
    description="{{.Description}}",
    schedule_interval="{{.Schedule}}",
    start_date=datetime.now() - timedelta(days=1),  # Start in the past
    catchup=True,
    tags=["dynamic"],
) as dag:
    start = DummyOperator(task_id="start")
    {{range .Tasks}}
    {{.TaskID}} = PythonOperator(
        task_id="{{.TaskID}}",
        python_callable=task_function,
        op_args=["{{.TaskID}}"],
    )
    start >> {{.TaskID}}
    {{end}}
`


// Main function
func main() {
	checkAirflowHealth()

    r := mux.NewRouter()
    r.HandleFunc("/create-workflow", createUserWorkflow).Methods("POST")
	r.HandleFunc("/trigger_spark", triggerSparkJob).Methods("POST")
    r.Handle("/metrics", promhttp.Handler())

	// Endpoints
	r.HandleFunc("/api/workflow", createWorkflow).Methods("POST")          // Create Workflow
	r.HandleFunc("/api/airflow/trigger", triggerAirflowDag).Methods("POST") // Trigger Airflow DAG
	r.HandleFunc("/api/spark/submit", submitSparkJob).Methods("POST")       // Submit Spark Job


	r.HandleFunc("/submit-workflow", submitWorkflow).Methods("POST")


	// Prometheus metrics
	// initPrometheus()

    log.Println("HTTP server started on :8082")
    log.Fatal(http.ListenAndServe(":8082", r))
}
