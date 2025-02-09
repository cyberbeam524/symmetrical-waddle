package main

import (
	// "bytes"
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/gorilla/mux"
	"github.com/playwright-community/playwright-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"services/scraper/controllers"
	"services/scraper/config"
	"github.com/joho/godotenv"
    "github.com/gorilla/handlers"
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

func generateTopicNames(taskTypes map[string]int) []string {
    var topicNames []string

    for taskType, count := range taskTypes {
        for i := 1; i <= count; i++ {
            topicNames = append(topicNames, fmt.Sprintf("%s_%d", taskType, i))
        }
    }

    return topicNames
}


func initKafka() {
    kafkaWriter = &kafka.Writer{
        Addr:     kafka.TCP("localhost:9092"),
        Balancer: &kafka.LeastBytes{},
    }
    log.Println("Kafka initialized")

    // Define the number of topics to create for each task type
    taskCounts := map[string]int{
        "scrape":    20, // Increase to 20 scrape topics
        "transform": 10,
        "store":     5,
    }    

    // Generate topic names dynamically
    topicNames := generateTopicNames(taskCounts)

    // Create topics
    for _, topic := range topicNames {
        if err := controllers.CreateKafkaTopic(topic); err != nil {
            log.Printf("Failed to create topic %s: %v", topic, err)
        } else {
            log.Printf("Topic created successfully: %s", topic)
        }
    }
}

func initMongo() {
    mongoURI := os.Getenv("MONGO_URI")
    log.Println("MongoDB initialized: %s", mongoURI)
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017" // Default fallback
    }

    var err error
    mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
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

// func triggerSparkJob(w http.ResponseWriter, r *http.Request) {
//     jobData := map[string]string{
//         "job_type": "spark_transformation",
//     }
//     jobDataJSON, _ := json.Marshal(jobData)
//     err := controllers.ProduceToKafka("spark_jobs", "spark_task", jobDataJSON)
//     if err != nil {
//         http.Error(w, "Failed to trigger Spark job", http.StatusInternalServerError)
//         return
//     }
//     w.WriteHeader(http.StatusOK)
//     w.Write([]byte("Spark job triggered successfully"))
// }


// // ---------------- Spark Integration ----------------

// // Submit a Spark Job
// func submitSparkJob(w http.ResponseWriter, r *http.Request) {
// 	type SparkJobPayload struct {
// 		JobName   string   `json:"job_name"`
// 		Arguments []string `json:"arguments"`
// 	}

// 	var payload SparkJobPayload
// 	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	sparkAPI := "http://spark-master:8083/v1/submissions/create"
// 	reqBody, _ := json.Marshal(map[string]interface{}{
// 		"action": "CreateSubmissionRequest",
// 		"appArgs": payload.Arguments,
// 		"appResource": fmt.Sprintf("local:/opt/spark-jobs/%s", payload.JobName),
// 	})

// 	resp, err := http.Post(sparkAPI, "application/json", bytes.NewBuffer(reqBody))
// 	if err != nil || resp.StatusCode != http.StatusOK {
// 		log.Printf("Failed to submit Spark job: %v", err)
// 		http.Error(w, "Failed to submit Spark job", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Spark job submitted successfully"))
// }

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

func seedTopics() {
    collection := mongoClient.Database("workflow_resources").Collection("topics")
    taskCounts := map[string]int{
        "scrape":    10,
        "transform": 5,
        "store":     3,
    }

    for taskType, count := range taskCounts {
        for i := 1; i <= count; i++ {
            topicName := fmt.Sprintf("%s_%d", taskType, i)

            // Insert topic into MongoDB
            _, err := collection.InsertOne(context.Background(), bson.M{
                "topic_name": topicName,
                "task_type":  taskType,
                "assigned":   false,
                "workflow_id": "",
            })
            if err != nil {
                log.Printf("Failed to insert topic '%s': %v", topicName, err)
            }
        }
    }

    log.Println("Topics seeded successfully")
}


func initResourcePool() error {
    collection := mongoClient.Database("resource_management").Collection("resources")
    topics := []string{"topic_scrape", "topic_transform", "topic_store"} // Example topics

    for _, topic := range topics {
        _, err := collection.UpdateOne(
            context.Background(),
            bson.M{"topic": topic},
            bson.M{
                "$setOnInsert": bson.M{
                    "topic":     topic,
                    "task_type": strings.Split(topic, "_")[1],
                    "status":    "available",
                    "workflow_id": nil,
                },
            },
            options.Update().SetUpsert(true),
        )
        if err != nil {
            log.Printf("Failed to initialize resource: %v", err)
            return err
        }
    }

    log.Println("Resource pool initialized")
    return nil
}


// Main function
func main() {

    err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Setup MongoDB, Stripe, and other configs
	config.Setup()

    errPlaywrightInstallation := playwright.Install()
    if errPlaywrightInstallation != nil{
        log.Println("playwright not installed")
    }
	controllers.CheckAirflowHealth()
    // initMongo()
    seedTopics()
	initResourcePool()

    r := mux.NewRouter()
    r.HandleFunc("/create-workflow", controllers.CreateUserWorkflow).Methods("POST")
	r.HandleFunc("/trigger_spark", controllers.TriggerSparkJob).Methods("POST")
    r.Handle("/metrics", promhttp.Handler())

	// Endpoints
	r.HandleFunc("/api/workflow", controllers.CreateWorkflow).Methods("POST")          // Create Workflow
	r.HandleFunc("/api/airflow/trigger", controllers.TriggerAirflowDag).Methods("POST") // Trigger Airflow DAG
	r.HandleFunc("/api/spark/submit", controllers.SubmitSparkJob).Methods("POST")       // Submit Spark Job


	r.HandleFunc("/submit-workflow", controllers.SubmitWorkflow).Methods("POST")

	r.HandleFunc("/scrape", controllers.ScrapeHandler).Methods("POST")

    // Resource management handlers for Airflow
    r.HandleFunc("/assign-resources", controllers.AssignResourcesHandler).Methods("POST")  // Assign resources for tasks
    r.HandleFunc("/release-resources", controllers.ReleaseResourcesHandler).Methods("POST") // Release resources after task completion


    // trino view and dashboard creation:
    r.HandleFunc("/create-trino-view", controllers.CreateTrinoView).Methods("POST") //
    // r.HandleFunc("/configure-superset-dataset", configureSupersetDataset).Methods("POST") // 
    // Replace r.HandleFunc with:
    r.HandleFunc("/configure-superset-dataset", controllers.HandleConfigureSupersetDataset)


    // Authentication routes
    r.HandleFunc("/auth/google", controllers.GoogleLogin).Methods("POST") //
    r.HandleFunc("/auth/email", controllers.EmailLogin).Methods("POST") //
    r.HandleFunc("/payment/checkout", controllers.CreateCheckoutSession).Methods("POST") //
    r.HandleFunc("/payment/success", controllers.HandlePaymentSuccess).Methods("GET") //
    r.HandleFunc("/payment/cancel", controllers.CancelSubscription).Methods("POST") //
    r.HandleFunc("/user/credits", controllers.GetUserCredits).Methods("GET")

    r.HandleFunc("/webhook/stripe", controllers.StripeWebhookHandler).Methods("POST")

	// Prometheus metrics handler
    r.Handle("/metrics", promhttp.Handler())

    // Define allowed headers, origins, and methods for CORS
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"http://localhost:3000"}), // Frontend origin
	)

	// Wrap your router with CORS middleware
	log.Println("Starting server on :8083")
	log.Fatal(http.ListenAndServe(":8083", cors(r)))
}
