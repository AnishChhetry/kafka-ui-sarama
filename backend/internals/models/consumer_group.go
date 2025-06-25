package models

// ConsumerGroup represents a Kafka consumer group and its metadata.
type ConsumerGroup struct {
	GroupID    string   `json:"groupId"`    // Consumer group ID
	MemberID   string   `json:"memberId"`   // Member ID
	Topics     []string `json:"topics"`     // Subscribed topics
	Partitions []int32  `json:"partitions"` // Assigned partitions
	Error      string   `json:"error"`      // Error message, if any
}
