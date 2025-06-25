package models

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
