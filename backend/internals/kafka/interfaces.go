package kafka

import (
	"backend/internals/models"
)

// interfaces.go - Defines interfaces and data structures for Kafka operations.
// Provides the KafkaService interface and related types for topics, partitions, brokers, consumers, and messages.

// KafkaService defines the interface for all Kafka operations, including connection, topic, message, and cluster management.
type KafkaService interface {
	// Connection Operations
	CheckConnection() error // Checks connectivity to the Kafka cluster

	// Topic Operations
	ListTopics() ([]models.Topic, error)                              // Lists all topics
	CreateTopic(name string, partitions, replicationFactor int) error // Creates a new topic
	DeleteTopic(topic string) error                                   // Deletes a topic
	GetPartitionInfo(topic string) ([]models.PartitionInfo, error)    // Gets partition info for a topic

	// Message Operations
	ClearTopicMessages(topic string) error                                                          // Clears all messages from a topic
	FetchMessages(topic string, limit int, sortOrder string) ([]models.Message, error)              // Fetches messages from a topic
	Produce(topic, key string, value []byte, partition int32, headers []models.MessageHeader) error // Produces a message

	// Cluster Operations
	GetBrokers() ([]models.Broker, error)          // Gets broker info
	GetConsumers() ([]models.ConsumerGroup, error) // Gets consumer group info
}
