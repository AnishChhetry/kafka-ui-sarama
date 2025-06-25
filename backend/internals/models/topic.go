package models

// Topic represents a Kafka topic, including partitions and consumer groups.
type Topic struct {
	Name              string          `json:"name"`              // Topic name
	Partitions        []Partition     `json:"partitions"`        // Partitions in the topic
	ConsumerGroups    []ConsumerGroup `json:"consumerGroups"`    // Consumer groups for the topic
	Internal          bool            `json:"internal"`          // Whether the topic is internal
	PartitionCount    int             `json:"partitionCount"`    // Number of partitions
	ReplicationFactor int             `json:"replicationFactor"` // Replication factor
}
