package kafka

// interfaces.go - Defines interfaces and data structures for Kafka operations.
// Provides the KafkaService interface and related types for topics, partitions, brokers, consumers, and messages.

// KafkaService defines the interface for all Kafka operations, including connection, topic, message, and cluster management.
type KafkaService interface {
	// Connection Operations
	CheckConnection() error // Checks connectivity to the Kafka cluster

	// Topic Operations
	ListTopics() ([]Topic, error)                                     // Lists all topics
	CreateTopic(name string, partitions, replicationFactor int) error // Creates a new topic
	DeleteTopic(topic string) error                                   // Deletes a topic
	GetPartitionInfo(topic string) ([]PartitionInfo, error)           // Gets partition info for a topic

	// Message Operations
	ClearTopicMessages(topic string) error                                                   // Clears all messages from a topic
	FetchMessages(topic string, limit int, sortOrder string) ([]Message, error)              // Fetches messages from a topic
	Produce(topic, key string, value []byte, partition int32, headers []MessageHeader) error // Produces a message

	// Cluster Operations
	GetBrokers() ([]BrokerInfo, error)          // Gets broker info
	GetConsumers() ([]ConsumerGroupInfo, error) // Gets consumer group info
}

// Message represents a Kafka message, including metadata and headers.
type Message struct {
	Topic     string          `json:"topic"`     // Topic name
	Partition int32           `json:"partition"` // Partition number
	Offset    int64           `json:"offset"`    // Message offset
	Key       string          `json:"key"`       // Message key
	Value     string          `json:"value"`     // Message value
	Timestamp int64           `json:"timestamp"` // Unix timestamp (ms)
	Headers   []MessageHeader `json:"headers"`   // Message headers
	Size      int             `json:"size"`      // Message size in bytes
}

// MessageHeader represents a Kafka message header (key-value pair).
type MessageHeader struct {
	Key   string `json:"key"`   // Header key
	Value string `json:"value"` // Header value
}

// PartitionInfo represents information about a Kafka partition.
type PartitionInfo struct {
	Topic          string  `json:"topic"`          // Topic name
	Partition      int32   `json:"partition"`      // Partition number
	Leader         int32   `json:"leader"`         // Leader broker ID
	Replicas       []int32 `json:"replicas"`       // Replica broker IDs
	InSyncReplicas []int32 `json:"inSyncReplicas"` // In-sync replica broker IDs
}

// Broker represents a Kafka broker and its metadata.
type Broker struct {
	ID           int32  `json:"id"`           // Broker ID
	Host         string `json:"host"`         // Hostname
	Port         int32  `json:"port"`         // Port number
	Address      string `json:"address"`      // Full address
	Status       string `json:"status"`       // Broker status
	SegmentCount int    `json:"segmentCount"` // Number of log segments
	Replicas     []int  `json:"replicas"`     // Replica partitions
	Leaders      []int  `json:"leaders"`      // Leader partitions
}

// ConsumerGroup represents a Kafka consumer group and its metadata.
type ConsumerGroup struct {
	GroupID    string   `json:"groupId"`    // Consumer group ID
	MemberID   string   `json:"memberId"`   // Member ID
	Topics     []string `json:"topics"`     // Subscribed topics
	Partitions []int32  `json:"partitions"` // Assigned partitions
	Error      string   `json:"error"`      // Error message, if any
}

// Topic represents a Kafka topic, including partitions and consumer groups.
type Topic struct {
	Name              string          `json:"name"`              // Topic name
	Partitions        []Partition     `json:"partitions"`        // Partitions in the topic
	ConsumerGroups    []ConsumerGroup `json:"consumerGroups"`    // Consumer groups for the topic
	Internal          bool            `json:"internal"`          // Whether the topic is internal
	PartitionCount    int             `json:"partitionCount"`    // Number of partitions
	ReplicationFactor int             `json:"replicationFactor"` // Replication factor
}

// Partition represents a Kafka topic partition and its metadata.
type Partition struct {
	ID              int   `json:"id"`              // Partition ID
	Leader          int   `json:"leader"`          // Leader broker ID
	Replicas        []int `json:"replicas"`        // Replica broker IDs
	InSyncReplicas  []int `json:"inSyncReplicas"`  // In-sync replica broker IDs
	OfflineReplicas []int `json:"offlineReplicas"` // Offline replica broker IDs
}

// BrokerInfo is an alias for Broker for interface compatibility.
type BrokerInfo = Broker

// ConsumerGroupInfo is an alias for ConsumerGroup for interface compatibility.
type ConsumerGroupInfo = ConsumerGroup
