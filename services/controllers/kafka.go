package controllers

import (
	"encoding/json"
	"context"
	"fmt"
	"log"
	"github.com/segmentio/kafka-go"
)

func incrementTaskCounter(taskType string) {
    taskProcessedCounter.WithLabelValues(taskType).Inc()
}

func CreateKafkaTopic(topic string) error {
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




func startWorkflowConsumers(userID string) {
    topic := fmt.Sprintf("workflow_%s", userID)
    startConsumer(topic, "transform", handleTransformTask)
    startConsumer(topic, "store", handleStoreTask)
}


// ---------------- Kafka Integration ----------------

func ProduceToKafka(topic string, taskType string, message interface{}) error {
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


func produceBatchToKafka(topic string, messages []map[string]interface{}) error {
	// Convert results to Kafka messages
	var kafkaMessages []kafka.Message
	for _, result := range messages {
        // Flatten the message
		flatMessage := flattenJSON(result)

		// Convert the flat message to JSON
		value, err := json.Marshal(flatMessage)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %v", err)
		}
		kafkaMessages = append(kafkaMessages, kafka.Message{
			Value: value,
		})
	}

	// Initialize Kafka writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	// Write messages in batch
	err := writer.WriteMessages(context.Background(), kafkaMessages...)
	if err != nil {
		log.Printf("Failed to write batch to Kafka: %v", err)
		return err
	}

	log.Printf("Batch of %d messages successfully written to Kafka topic: %s", len(kafkaMessages), topic)
	return nil
}