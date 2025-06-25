package models

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
