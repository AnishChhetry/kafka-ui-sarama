// topic.go - Defines the Topic model for representing Kafka topics in the application.

// Topic represents a Kafka topic with its name, partitions, and replicas.
package models

type Topic struct {
	Name       string   `json:"name"`
	Partitions []int    `json:"partitions"`
	Replicas   []string `json:"replicas"`
}
