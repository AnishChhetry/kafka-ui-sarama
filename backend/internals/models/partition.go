package models

// PartitionInfo represents information about a Kafka partition.
type PartitionInfo struct {
	Topic          string  `json:"topic"`          // Topic name
	Partition      int32   `json:"partition"`      // Partition number
	Leader         int32   `json:"leader"`         // Leader broker ID
	Replicas       []int32 `json:"replicas"`       // Replica broker IDs
	InSyncReplicas []int32 `json:"inSyncReplicas"` // In-sync replica broker IDs
}

// Partition represents a Kafka topic partition and its metadata.
type Partition struct {
	ID              int   `json:"id"`              // Partition ID
	Leader          int   `json:"leader"`          // Leader broker ID
	Replicas        []int `json:"replicas"`        // Replica broker IDs
	InSyncReplicas  []int `json:"inSyncReplicas"`  // In-sync replica broker IDs
	OfflineReplicas []int `json:"offlineReplicas"` // Offline replica broker IDs
}
