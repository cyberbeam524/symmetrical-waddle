package controllers

import (

	"encoding/json"
	"context"
	"fmt"
	"net/http"
	"log"
	"io"
	"bytes"
    "text/template"
    "os"
	"os/exec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"

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

func assignResourcesForWorkflow(workflowID string, tasks []Task) (map[string]string, error) {
    topicAssignments := make(map[string]string)

    collection := mongoClient.Database("workflow_resources").Collection("topics")

    for _, task := range tasks {
        filter := bson.M{
            "task_type": task.TaskType,
            "assigned":  false,
        }

        update := bson.M{
            "$set": bson.M{
                "assigned":    true,
                "workflow_id": workflowID,
            },
        }

        // Find and update an available topic
        result := collection.FindOneAndUpdate(context.Background(), filter, update)
        if result.Err() != nil {
            return nil, fmt.Errorf("no available topic for task type: %s", task.TaskType)
        }

        var topicData struct {
            TopicName string `bson:"topic_name"`
        }
        if err := result.Decode(&topicData); err != nil {
            return nil, fmt.Errorf("failed to decode topic data: %v", err)
        }

        topicAssignments[task.TaskID] = topicData.TopicName
    }

    log.Printf("Resources acquired for workflow '%s'", workflowID)

    return topicAssignments, nil
}


func releaseResources(workflowID string) error {
    collection := mongoClient.Database("workflow_resources").Collection("topics")

    _, err := collection.UpdateMany(
        context.Background(),
        bson.M{"workflow_id": workflowID},
        bson.M{"$set": bson.M{"assigned": false, "workflow_id": ""}},
    )
    if err != nil {
        return fmt.Errorf("failed to release resources: %v", err)
    }

    log.Printf("Resources released for workflow '%s'", workflowID)
    return nil
}



// topic creation and reassignments:
func AssignResourcesHandler(w http.ResponseWriter, r *http.Request) {
    var workflow Workflow
    if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    taskTopicMap, err := assignResourcesForWorkflow(workflow.WorkflowID, workflow.Tasks)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return assigned topics and consumer groups
    json.NewEncoder(w).Encode(taskTopicMap)
}


func ReleaseResourcesHandler(w http.ResponseWriter, r *http.Request) {
    workflowID := r.URL.Query().Get("workflow_id")
    if workflowID == "" {
        http.Error(w, "Missing workflow_id", http.StatusBadRequest)
        return
    }

    if err := releaseResources(workflowID); err != nil {
        http.Error(w, fmt.Sprintf("Failed to release resources: %v", err), http.StatusInternalServerError)
        return
    }

    log.Printf("Workflow '%s' resources released successfully", workflowID)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Resources released successfully"))
}


// ---------------- Airflow Integration ----------------

// Trigger an Airflow DAG
func TriggerAirflowDag(w http.ResponseWriter, r *http.Request) {
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


func SubmitWorkflow(w http.ResponseWriter, r *http.Request) {
    var workflow Workflow
    if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    if workflow.UserID == "" || workflow.WorkflowID == "" {
        http.Error(w, "Missing required fields: user_id or workflow_id", http.StatusBadRequest)
        return
    }

    // // Assign resources for the workflow
    // taskTopicMap, err := assignResourcesForWorkflow(workflow.WorkflowID, workflow.Tasks)
    // if err != nil {
    //     http.Error(w, fmt.Sprintf("Failed to assign resources: %v", err), http.StatusInternalServerError)
    //     return
    // }

    // Trigger DAG creation in Airflow
    if err := triggerAirflowDagCreation(workflow); err != nil {
        http.Error(w, "Failed to trigger Airflow DAG creation", http.StatusInternalServerError)
        return
    }

    log.Printf("Workflow '%s' submitted successfully with resources", workflow.WorkflowID)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Workflow submitted successfully"))
}


// // Endpoint to submit a workflow
// func submitWorkflow_older(w http.ResponseWriter, r *http.Request) {
// 	var workflow Workflow
// 	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

//     // Validate the workflow payload
// 	if workflow.UserID == "" || workflow.WorkflowID == "" {
// 		http.Error(w, "Missing required fields: user_id or workflow_id", http.StatusBadRequest)
// 		return
// 	}

// 	if workflow.Tasks == nil || len(workflow.Tasks) == 0 {
// 		http.Error(w, "Workflow must contain at least one task", http.StatusBadRequest)
// 		return
// 	}

// 	// Save workflow configuration to file
// 	filePath := filepath.Join("workflows", fmt.Sprintf("%s_%s.json", workflow.UserID, workflow.WorkflowID))
// 	if err := os.MkdirAll("workflows", os.ModePerm); err != nil {
// 		http.Error(w, "Failed to create directory for workflows", http.StatusInternalServerError)
// 		return
// 	}
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()
// 	if err := json.NewEncoder(file).Encode(workflow); err != nil {
// 		http.Error(w, "Failed to save workflow", http.StatusInternalServerError)
// 		return
// 	}

// 	// Trigger DAG creation in Airflow
// 	if err := triggerAirflowDagCreation(workflow); err != nil {
// 		http.Error(w, "Failed to trigger Airflow DAG creation", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Workflow submitted successfully"))
// }



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
        WorkflowID  string
		Tasks       []struct {
			TaskID string
            DependsOn []string
		}
	}{
		DagID:       fmt.Sprintf("workflow_%s_%s", workflow.UserID, workflow.WorkflowID),
		Description: "User-defined workflow DAG",
        Schedule:    workflow.Schedule, // User-defined schedule
        WorkflowID:  workflow.WorkflowID,
	}

	// Add tasks to the DAG
	for _, task := range workflow.Tasks {
		dagData.Tasks = append(dagData.Tasks, struct {
			TaskID string
            DependsOn []string
		}{
			TaskID: task.TaskID,
            DependsOn: task.DependsOn, // Ensure DependsOn is correctly populated from the workflow
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


func CheckAirflowHealth() error {
    resp, err := http.Get("http://localhost:8081/health")
    if err != nil {
        return fmt.Errorf("Airflow not reachable: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("Airflow health check failed: %d", resp.StatusCode)
    }
    return nil
}


const airflowDagTemplate = `
from airflow import DAG
from airflow.operators.python import PythonOperator
from airflow.operators.dummy import DummyOperator
from datetime import datetime, timedelta
import requests
import logging
from kafka import KafkaConsumer
import json


logger = logging.getLogger(__name__)
logger.info("This is a log message")

def consume_kafka_messages(topic, group_id, bootstrap_servers, process_message):
    """
    Consumes messages from Kafka and processes them.
    """
    consumer = KafkaConsumer(
        topic,
        group_id=group_id,
        bootstrap_servers=bootstrap_servers,
        auto_offset_reset='earliest',
        enable_auto_commit=False,  # Explicit commit
        value_deserializer=lambda x: json.loads(x.decode('utf-8'))
    )

    logging.info(f"Consuming messages from topic: {topic}")
    
    for message in consumer:
        logging.info(f"Consumed message: {message.value}")
        
        # Process the message
        process_message(message.value)
        
        # Commit the message offset
        consumer.commit()

    logging.info(f"Finished consuming messages from topic: {topic}")
    consumer.close()


def consume_and_process_kafka(**kwargs):
    topic = kwargs['topic']
    group_id = kwargs['group_id']
    bootstrap_servers = kwargs['bootstrap_servers']
    
    def process_message(message):
        # Add transformation, deduplication, or storage logic here
        transformed_message = {key: value.upper() for key, value in message.items()}
        print(f"Processed message: {transformed_message}")

    consume_kafka_messages(
        topic=topic,
        group_id=group_id,
        bootstrap_servers=bootstrap_servers,
        process_message=process_message
    )

def assign_resources(task_id, workflow_id):
    url = "http://host.docker.internal:8082/assign-resources"
    payload = {
        "workflow_id": workflow_id,
        "tasks": [
         {{ range $index, $task := .Tasks }}
        {
            "task_id": "{{$task.TaskID}}",
            "task_type": "{{$task.TaskID}}",
        },
         {{ end }}
        ]
    }
    response = requests.post(url, json=payload)
    response.raise_for_status()
    return response.json()

def release_resources(workflow_id):
    url = f"http://host.docker.internal:8082/release-resources?workflow_id={workflow_id}"
    response = requests.post(url)
    response.raise_for_status()

def execute_task(task_id, workflow_id):
    # Request resource assignment
    topic_map = assign_resources(task_id, workflow_id)
    print(f"Resources assigned for task {task_id}: {topic_map}")

    try:
        # Simulate task execution
        print(f"Executing task {task_id} for workflow {workflow_id}")
        logger.info(f"Executing task {task_id} for workflow {workflow_id}")
    finally:
        # Release resources after execution
        release_resources(workflow_id)
        logger.info("Finish release_resources")

default_args = {
    'owner': 'airflow',
    'depends_on_past': False,
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

with DAG(
    dag_id="{{ .DagID }}",
    default_args=default_args,
    description="{{ .Description }}",
    schedule_interval="{{ .Schedule }}",
    start_date=datetime.now() - timedelta(days=1),
    catchup=True,
    tags=["dynamic"],
) as dag:
    start = DummyOperator(task_id="start")
    end = DummyOperator(task_id="end")
    workflow_id = "{{ .WorkflowID }}"

    # Dictionary to store created tasks
    task_dict = {}

    {{ range $index, $task := .Tasks }}
    task_dict["{{ $task.TaskID }}"] = PythonOperator(
        task_id="{{ $task.TaskID }}",
        python_callable=execute_task if "{{.TaskID}}" != "store" else consume_and_process_kafka,
        op_args=["{{ $task.TaskID }}", workflow_id] if "{{.TaskID}}" != "store" else {
            'topic': 'store-1',
            'group_id': 'example-group',
            'bootstrap_servers': ['localhost:9092']
        },
    )
    {{ end }}

    {{ range $index, $task := .Tasks }}
    {{if not $task.DependsOn }}
    start >> task_dict["{{ $task.TaskID }}"]
    {{ else }}
    {{ range $index, $dependency := $task.DependsOn }}
    task_dict["{{ $dependency }}"] >> task_dict["{{ $task.TaskID }}"]
    {{ end }}
    {{ end }}
    {{ end }}

    terminal_tasks = list(task_dict.keys())
    {{ range $index, $task := .Tasks }}
    {{ range $index, $dependency := $task.DependsOn }}
    terminal_tasks.remove("{{$dependency}}")
    {{ end }}
    {{ end }}

    for task_id, task in task_dict.items():
        if task_id in terminal_tasks:
            task >> end
`


// const airflowDagTemplate3 = `
// from airflow import DAG
// from airflow.operators.python import PythonOperator
// from airflow.operators.dummy import DummyOperator
// from datetime import datetime, timedelta
// import requests
// import logging

// logger = logging.getLogger(__name__)
// logger.info("This is a log message")

// def assign_resources(task_id, workflow_id):
//     url = "http://localhost:8082/assign-resources"
//     payload = {
//         "workflow_id": workflow_id,
//         "tasks": [{"task_id": task_id}]
//     }
//     response = requests.post(url, json=payload)
//     response.raise_for_status()
//     return response.json()

// def release_resources(workflow_id):
//     url = f"http://localhost:8082/release-resources?workflow_id={workflow_id}"
//     response = requests.post(url)
//     response.raise_for_status()

// def execute_task(task_id, workflow_id):
//     # Request resource assignment
//     topic_map = assign_resources(task_id, workflow_id)
//     print(f"Resources assigned for task {task_id}: {topic_map}")

//     try:
//         # Simulate task execution
//         print(f"Executing task {task_id} for workflow {workflow_id}")
//         logger.info(f"Executing task {task_id} for workflow {workflow_id}")
//     finally:
//         # Release resources after execution
//         release_resources(workflow_id)
//         logger.info("Finish release_resources")

// default_args = {
//     'owner': 'airflow',
//     'depends_on_past': False,
//     'email_on_failure': False,
//     'email_on_retry': False,
//     'retries': 1,
//     'retry_delay': timedelta(minutes=5),
// }

// def task_function(task_id):
//     print(f"Executing task: {task_id}")

// with DAG(
//     dag_id="{{.DagID}}",
//     default_args=default_args,
//     description="{{.Description}}",
//     schedule_interval="{{.Schedule}}",
//     start_date=datetime.now() - timedelta(days=1),  # Start in the past
//     catchup=True,
//     tags=["dynamic"],
// ) as dag:
//     start = DummyOperator(task_id="start")

//     workflow_id = "{{.WorkflowID}}"  # Make sure WorkflowID is passed correctly

//     {{range .Tasks}}
//     {{.TaskID}} = PythonOperator(
//         task_id="{{.TaskID}}",
//         python_callable=task_function,
//         op_args=["{{.TaskID}}", workflow_id],
//     )
//     start >> {{.TaskID}}
//     {{end}}
//     `


// // Template for the dynamic Airflow DAG
// const airflowDagTemplate2 = `
// from airflow import DAG
// from airflow.operators.dummy import DummyOperator
// from airflow.operators.python import PythonOperator
// from datetime import datetime, timedelta

// import requests

// def assign_resources(workflow_id, tasks):
//     url = "http://host.docker.internal:8082/assign-resources"
//     payload = {
//         "workflow_id": workflow_id,
//         "tasks": tasks
//     }
//     response = requests.post(url, json=payload)
//     response.raise_for_status()
//     return response.json()

// def release_resources(workflow_id):
//     url = f"http://host.docker.internal:8082/release-resources?workflow_id={workflow_id}"
//     response = requests.post(url)
//     response.raise_for_status()

// # Example usage in Airflow tasks
// def airflow_task_assign_resources():
//     workflow_id = "workflow_user1"
//     tasks = [{"task_id": "scrape"}, {"task_id": "transform"}]
//     topic_map = assign_resources(workflow_id, tasks)
//     print("Assigned resources:", topic_map)

// def airflow_task_release_resources():
//     workflow_id = "workflow_user1"
//     release_resources(workflow_id)
//     print(f"Released resources for workflow {workflow_id}")


// default_args = {
//     'owner': 'airflow',
//     'depends_on_past': False,
//     'email_on_failure': False,
//     'email_on_retry': False,
//     'retries': 1,
//     'retry_delay': timedelta(minutes=5),
// }

// def task_function(task_id):
//     print(f"Executing task: {task_id}")

// with DAG(
//     dag_id="{{.DagID}}",
//     default_args=default_args,
//     description="{{.Description}}",
//     schedule_interval="{{.Schedule}}",
//     start_date=datetime.now() - timedelta(days=1),  # Start in the past
//     catchup=True,
//     tags=["dynamic"],
// ) as dag:
//     start = DummyOperator(task_id="start")
//     {{range .Tasks}}
//     {{.TaskID}} = PythonOperator(
//         task_id="{{.TaskID}}",
//         python_callable=task_function,
//         op_args=["{{.TaskID}}"],
//     )
//     start >> {{.TaskID}}
//     {{end}}
// `