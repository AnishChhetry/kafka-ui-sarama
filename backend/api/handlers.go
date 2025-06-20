package api

import (
	"backend/kafka"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// handlers.go - Contains HTTP handler functions for Kafka-related API endpoints.
// Provides endpoints for managing topics, messages, brokers, consumers, and connection checks.

var kafkaService kafka.KafkaService

// Initialize sets the Kafka service instance for use by all API handlers.
func Initialize(service kafka.KafkaService) {
	kafkaService = service
}

// GetTopics returns a list of all Kafka topics.
// Response: 200 OK with JSON array of topics, or 500 Internal Server Error.
func GetTopics(c *gin.Context) {
	topics, err := kafkaService.ListTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, topics)
}

// GetMessages fetches messages from a given topic.
// Query params:
//   - limit: number of messages to fetch (default 5)
//   - sort: 'newest' or 'oldest' (default 'newest')
//
// Response: 200 OK with JSON array of messages, or 500 Internal Server Error.
func GetMessages(c *gin.Context) {
	topic := c.Param("name")
	limitStr := c.DefaultQuery("limit", "5")
	limit, _ := strconv.Atoi(limitStr)
	sortOrder := c.DefaultQuery("sort", "newest")

	messages, err := kafkaService.FetchMessages(topic, limit, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// ProduceMessage produces a message to a Kafka topic.
// Request JSON body:
//
//	{
//	  "topic": "<topic>",
//	  "key": "<key>",
//	  "value": "<value>",
//	  "partition": <partition>,
//	  "headers": [ { "key": "<key>", "value": "<value>" }, ... ]
//	}
//
// Response: 200 OK on success, 400 Bad Request or 500 Internal Server Error on failure.
func ProduceMessage(c *gin.Context) {
	type reqBody struct {
		Topic     string `json:"topic"`
		Key       string `json:"key,omitempty"`
		Value     string `json:"value,omitempty"`
		Partition int32  `json:"partition"`
		Headers   []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"headers,omitempty"`
	}
	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// If partition is -1, let Kafka choose the partition
	var partition int32 = -1
	if body.Partition >= 0 {
		partition = body.Partition
	}

	// Convert headers to Kafka MessageHeader type
	headers := make([]kafka.MessageHeader, len(body.Headers))
	for i, h := range body.Headers {
		headers[i] = kafka.MessageHeader{
			Key:   h.Key,
			Value: h.Value,
		}
	}

	if err := kafkaService.Produce(body.Topic, body.Key, []byte(body.Value), partition, headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// DeleteMessages clears all messages from a given topic.
// Response: 200 OK on success, 500 Internal Server Error on failure.
func DeleteMessages(c *gin.Context) {
	topic := c.Param("name")
	// Use improved message clearing method
	if err := kafkaService.ClearTopicMessages(topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// CreateTopic creates a new Kafka topic with the specified parameters.
// Request JSON body:
//
//	{
//	  "name": "<topic_name>",
//	  "partitions": <num_partitions>,
//	  "replicationFactor": <replication_factor>
//	}
//
// Response: 200 OK on success, 400 Bad Request or 500 Internal Server Error on failure.
func CreateTopic(c *gin.Context) {
	type reqBody struct {
		Name              string `json:"name"`
		Partitions        int    `json:"partitions"`
		ReplicationFactor int    `json:"replicationFactor"`
	}
	var body reqBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate input
	if body.Partitions < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Number of partitions must be at least 1"})
		return
	}

	if body.ReplicationFactor < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Replication factor must be at least 1"})
		return
	}

	if err := kafkaService.CreateTopic(body.Name, body.Partitions, body.ReplicationFactor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// GetPartitionInfo returns partition information for a given topic.
// Response: 200 OK with partition info, or 500 Internal Server Error.
func GetPartitionInfo(c *gin.Context) {
	topic := c.Param("name")
	partitions, err := kafkaService.GetPartitionInfo(topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, partitions)
}

// GetBrokers returns a list of Kafka brokers.
// Response: 200 OK with broker list, or 500 Internal Server Error.
func GetBrokers(c *gin.Context) {
	brokers, err := kafkaService.GetBrokers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, brokers)
}

// GetConsumers returns a list of Kafka consumers.
// Response: 200 OK with consumer list, or 500 Internal Server Error.
func GetConsumers(c *gin.Context) {
	consumers, err := kafkaService.GetConsumers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consumers)
}

// CheckConnection checks connectivity to a Kafka broker using the provided bootstrap server.
// Query param: bootstrapServer (required)
// Response: 200 OK on success, 400 Bad Request or 500 Internal Server Error on failure.
func CheckConnection(c *gin.Context) {
	bootstrapServer := c.Query("bootstrapServer")
	if bootstrapServer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bootstrapServer parameter is required"})
		return
	}

	brokers := []string{bootstrapServer}
	client, err := kafka.NewKafkaClient(brokers, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := client.CheckConnection(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	kafkaService = client
	c.JSON(http.StatusOK, gin.H{"status": "connected"})
}

// DeleteTopic deletes a Kafka topic by name.
// Response: 200 OK on success, 500 Internal Server Error on failure.
func DeleteTopic(c *gin.Context) {
	topic := c.Param("name")
	if err := kafkaService.DeleteTopic(topic); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
