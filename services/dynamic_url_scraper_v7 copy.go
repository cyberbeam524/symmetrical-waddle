package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"

    "github.com/PuerkitoBio/goquery"
    "github.com/gorilla/mux"
    _ "github.com/ClickHouse/clickhouse-go"
    "github.com/gorilla/handlers"
    "context"
	// "time"

	"github.com/segmentio/kafka-go"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

    "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)



var kafkaWriter *kafka.Writer

func initKafka() {
	// Initialize Kafka writer
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Balancer: &kafka.LeastBytes{},
	}
}

func createKafkaTopic(topic string) error {
	// Kafka topic creation logic
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


func produceToKafka(topic string, taskType string, message interface{}) error {
	// Serialize task metadata and payload
	msg, err := json.Marshal(map[string]interface{}{
		"task_type": taskType,
		"payload":   message,
	})
	if err != nil {
		return fmt.Errorf("Failed to serialize message: %v", err)
	}

	// Write message to Kafka
	err = kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(taskType),
			Value: msg,
		},
	)
	if err != nil {
		return fmt.Errorf("Failed to write message to Kafka: %v", err)
	}
	log.Printf("Message for task '%s' written to topic '%s'", taskType, topic)
	return nil
}

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

			// Route tasks based on task type
			if task["task_type"] == taskType {
				if err := handler(msg.Value); err != nil {
					log.Printf("Task handling failed: %v", err)
				}
			}
		}
	}()
}


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
}

func handleTransformTask(msg []byte) error {
	log.Printf("Transforming data: %s", string(msg))
	// Perform transformation logic
	transformedData := map[string]interface{}{"transformed": string(msg)}

	// Store transformed data in MongoDB
	return storeData("transformed_tasks", transformedData)
}

func handleStoreTask(msg []byte) error {
	log.Printf("Storing data: %s", string(msg))
	// Store final data in MongoDB
	return storeData("final_tasks", msg)
}

var taskProcessedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "task_processed_total",
		Help: "Total number of tasks processed by type",
	},
	[]string{"task_type"},
)

func initPrometheus() {
	prometheus.MustRegister(taskProcessedCounter)
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9091", nil)
}

func incrementTaskCounter(taskType string) {
	taskProcessedCounter.WithLabelValues(taskType).Inc()
}



func createUserPipeline(w http.ResponseWriter, r *http.Request) {
	type RequestPayload struct {
		UserID string `json:"user_id"`
		Tasks  []struct {
			TaskType string `json:"task_type"`
			Details  string `json:"details"`
		} `json:"tasks"`
	}

	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	topic := "pipeline_" + payload.UserID
	if err := createKafkaTopic(topic); err != nil {
		http.Error(w, "Failed to create Kafka topic", http.StatusInternalServerError)
		return
	}

	for _, task := range payload.Tasks {
		message, _ := json.Marshal(task)
		if err := produceToKafka(topic, "task", message); err != nil {
			http.Error(w, "Failed to enqueue tasks", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User pipeline created and tasks enqueued successfully"))
}




var (
	taskCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_processed_total",
			Help: "Total number of tasks processed",
		},
		[]string{"task_type"},
	)
)



func processTask(w http.ResponseWriter, r *http.Request) {
	var task struct {
		TaskType string `json:"task_type"`
		Details  string `json:"details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}

	// Process the task...
	incrementTaskCounter(task.TaskType)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task processed"))
}

var mongoClient *mongo.Client

func initMongo() {
	var err error
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")
}

func storeData(collectionName string, data interface{}) error {
	collection := mongoClient.Database("pipeline_data").Collection(collectionName)
	_, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		log.Printf("Failed to store data in MongoDB: %v", err)
	}
	return err
}


func startWorkflowConsumers(userID string) {
	topic := fmt.Sprintf("workflow_%s", userID)

	// Start consumers for transform and store tasks
	startConsumer(topic, "transform", handleTransformTask)
	startConsumer(topic, "store", handleStoreTask)
}




var (
    clickhouseConn *sql.DB
)

type FieldSelection struct {
    URL    string                         `json:"url"`
    Groups map[string]map[string]string   `json:"groups"` // Map of groups with their field selectors
	LCASelector string                        `json:"lcaSelector,omitempty"` // Optional direct LCA selector
}

func initClickhouse() {
    var err error
    clickhouseConn, err = sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
    if err != nil {
        log.Fatalf("Error connecting to ClickHouse: %v", err)
    }

    if err = clickhouseConn.Ping();
    
    err != nil {
        log.Fatalf("Could not ping ClickHouse: %v", err)
    }
}

func main() {
    // initClickhouse()
    r := mux.NewRouter()
    r.HandleFunc("/scrape", scrapeHandler).Methods("POST")

    corsObj := handlers.AllowedOrigins([]string{"*"})
    headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

    log.Println("HTTP server started on :8080")
    http.ListenAndServe(":8082", handlers.CORS(corsObj, headersOk, methodsOk)(r))
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("func scrapeHandler executing");
    var selection FieldSelection
    if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
        log.Println("Invalid request body");
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if groupFields, ok := selection.Groups["fields"]; ok {
        results, err := findAndExtractData(selection.URL, groupFields, selection.LCASelector)
        if err != nil {
            log.Printf("Error scraping URL %s: %v", selection.URL, err)
            http.Error(w, "Failed to scrape the page", http.StatusInternalServerError)
            return
        }
        log.Println(results);
        jsonData, err := json.Marshal(results)
        if err != nil {
            log.Println("Failed to serialize scraped data");
            http.Error(w, "Failed to serialize scraped data", http.StatusInternalServerError)
            return
        }

        log.Println("func scrapeHandler success");
        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonData)
    } else {
        log.Println("func scrapeHandler failed");
        http.Error(w, "No field group 'fields' found in request", http.StatusBadRequest)
    }
}

func findAndExtractData(url string, fields map[string]string, lcaSelector string) ([]map[string]interface{}, error) {
    doc, err := goquery.NewDocument(url)
    if err != nil {
        log.Printf("Error loading document: %v", err)
        return nil, err
    }

	// lcaSelector, err := findLowestCommonAncestorSelector(doc, fields)
	// if err != nil {
	// 	log.Printf("Error finding LCA selector: %v", err)
	// 	return nil, err
	// }

	// lcaSelector = "div.quote"
	// log.Printf("LCA Selector found: %s", lcaSelector)

    return extractDataFromAncestors2(doc, lcaSelector, fields), nil
}

func findLowestCommonAncestorSelector(doc *goquery.Document, fields map[string]string) (string, error) {
    if len(fields) == 0 {
        return "", fmt.Errorf("no fields provided")
    }

    type pathInfo struct {
        path  []*goquery.Selection
        depth int
    }

    // Map to store the paths of all elements
    elementPaths := make(map[string][]pathInfo)

    // Retrieve and store all paths for each field's elements
    for field, selector := range fields {
        elements := doc.Find(selector)
        if elements.Length() == 0 {
            return "", fmt.Errorf("no elements found for selector: %s", selector)
        }

        elements.Each(func(i int, s *goquery.Selection) {
            var path []*goquery.Selection
            for n := s; n.Length() > 0; n = n.Parent() {
                path = append([]*goquery.Selection{n}, path...)
            }
            elementPaths[field] = append(elementPaths[field], pathInfo{path: path, depth: len(path)})
        })
    }

    // Find the lowest common ancestor
    var lca *goquery.Selection
    minDepth := int(^uint(0) >> 1) // Initialize to max int

    // Initialize lca to the root of the first element's path
    for _, paths := range elementPaths {
        if len(paths) > 0 && paths[0].depth < minDepth {
            minDepth = paths[0].depth
            lca = paths[0].path[minDepth-1]
        }
    }

    // Compare all paths to find the common deepest element
    for _, paths := range elementPaths {
        for depth := 0; depth < minDepth; depth++ {
            current := paths[0].path[depth]
            allMatch := true
            for _, path := range paths {
                if path.path[depth].Get(0) != current.Get(0) {
                    allMatch = false
                    break
                }
            }
            if allMatch {
                lca = current
            } else {
                break
            }
        }
    }

    if lca == nil {
        return "", fmt.Errorf("no common ancestor found")
    }

    // Generate a unique selector for the LCA
    tag := goquery.NodeName(lca)
    id, exists := lca.Attr("id")
    if exists && id != "" {
        return tag + "#" + id, nil
    }
    classes, exists := lca.Attr("class")
    if exists && classes != "" {
        classList := strings.Split(classes, " ")
        return tag + "." + strings.Join(classList, "."), nil
    }
    return tag, nil
}


func extractDataFromAncestors2(doc *goquery.Document, lcaSelector string, fields map[string]string) []map[string]interface{} {
    var results []map[string]interface{}
    doc.Find(lcaSelector).Each(func(_ int, s *goquery.Selection) {
        result := make(map[string]interface{})
        for fieldName, selector := range fields {
            if text := s.Find(selector).Text(); text != "" {
                result[fieldName] = strings.TrimSpace(text)
            }
        }
        if len(result) > 0 {
            results = append(results, result)
        }
    })
    return results
}

func extractDataFromAncestors(doc *goquery.Document, lcaSelector string, fields map[string]string) []map[string]interface{} {
    var results []map[string]interface{}
    doc.Find(lcaSelector).Each(func(_ int, s *goquery.Selection) {
        result := make(map[string]interface{})
        for fieldName, fieldSelector := range fields {
            // Find and extract text specifically for each field within the LCA
            fieldValue := s.Find(fieldSelector).Text()
            if fieldValue != "" {
                result[fieldName] = strings.TrimSpace(fieldValue)
            }
        }
        if len(result) > 0 { // Ensure the result map is not empty before appending
            results = append(results, result)
        }
    })
    return results
}
