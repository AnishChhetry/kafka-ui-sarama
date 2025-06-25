package kafka

import (
	"backend/internals/models"
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// client.go - Implements KafkaService using IBM Sarama for Kafka operations.
// Provides methods for connecting to Kafka, managing topics, producing and consuming messages, and cluster operations.

// Client implements KafkaService using IBM Sarama client and admin interfaces.
type Client struct {
	brokers []string            // List of Kafka broker addresses
	config  *sarama.Config      // Sarama configuration
	client  sarama.Client       // Sarama client instance
	admin   sarama.ClusterAdmin // Sarama admin instance
}

// NewClient creates a new Kafka client using Sarama.
// brokers: list of broker addresses
// config: optional Sarama configuration (uses defaults if nil)
// Returns a pointer to Client and error if creation fails.
func NewClient(brokers []string, config *sarama.Config) (*Client, error) {
	if config == nil {
		config = sarama.NewConfig()
		config.Version = sarama.MaxVersion
		config.Producer.Return.Successes = true
		config.Producer.Partitioner = sarama.NewManualPartitioner

		// Enhanced settings for better admin operations
		config.Metadata.RefreshFrequency = 30 * time.Second
		config.Metadata.Full = true
		config.Admin.Timeout = 30 * time.Second
	}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return &Client{
		brokers: brokers,
		config:  config,
		client:  client,
		admin:   admin,
	}, nil
}

// NewKafkaClient is an alias for NewClient for compatibility.
func NewKafkaClient(brokers []string, config *sarama.Config) (*Client, error) {
	return NewClient(brokers, config)
}

// CheckConnection checks if the client can connect to the Kafka cluster.
// Returns error if no brokers are available or not connected.
func (c *Client) CheckConnection() error {
	brokers := c.client.Brokers()
	if len(brokers) == 0 {
		return errors.New("no brokers available")
	}
	for _, b := range brokers {
		if err := b.Open(c.config); err != nil && err != sarama.ErrAlreadyConnected {
			return err
		}
		connected, err := b.Connected()
		if err != nil || !connected {
			return errors.New("not connected to broker")
		}
	}
	return nil
}

// ListTopics lists all topics in the Kafka cluster.
// Returns a slice of Topic and error if listing fails.
func (c *Client) ListTopics() ([]models.Topic, error) {
	topicNames, err := c.client.Topics()
	if err != nil {
		return nil, err
	}
	details, err := c.admin.DescribeTopics(topicNames)
	if err != nil {
		return nil, err
	}
	var topics []models.Topic
	for _, meta := range details {
		partitionCount := len(meta.Partitions)
		replicationFactor := 0
		partitions := []models.Partition{}
		for _, p := range meta.Partitions {
			partitions = append(partitions, models.Partition{
				ID:              int(p.ID),
				Leader:          int(p.Leader),
				Replicas:        convertReplicas(p.Replicas),
				InSyncReplicas:  convertReplicas(p.Isr),
				OfflineReplicas: []int{}, // Sarama does not provide this
			})
		}
		if partitionCount > 0 {
			replicationFactor = len(meta.Partitions[0].Replicas)
		}
		topics = append(topics, models.Topic{
			Name:              meta.Name,
			Partitions:        partitions,
			ConsumerGroups:    []models.ConsumerGroup{}, // For now, empty
			Internal:          meta.IsInternal,
			PartitionCount:    partitionCount,
			ReplicationFactor: replicationFactor,
		})
	}
	return topics, nil
}

// CreateTopic creates a new topic with the given name, partitions, and replication factor.
// Returns error if creation fails.
func (c *Client) CreateTopic(name string, partitions, replicationFactor int) error {
	detail := &sarama.TopicDetail{
		NumPartitions:     int32(partitions),
		ReplicationFactor: int16(replicationFactor),
	}
	err := c.admin.CreateTopic(name, detail, false)
	if err == nil {
		_ = c.client.RefreshMetadata(name)
	}
	return err
}

// GetPartitionInfo gets partition info for a topic.
// Returns a slice of PartitionInfo and error if retrieval fails.
func (c *Client) GetPartitionInfo(topic string) ([]models.PartitionInfo, error) {
	meta, err := c.client.Partitions(topic)
	if err != nil {
		return nil, err
	}
	var infos []models.PartitionInfo
	for _, pid := range meta {
		leader, err := c.client.Leader(topic, pid)
		if err != nil {
			return nil, err
		}
		replicas, err := c.client.Replicas(topic, pid)
		if err != nil {
			return nil, err
		}
		isr, err := c.client.InSyncReplicas(topic, pid)
		if err != nil {
			return nil, err
		}
		infos = append(infos, models.PartitionInfo{
			Topic:          topic,
			Partition:      pid,
			Leader:         leader.ID(),
			Replicas:       replicas,
			InSyncReplicas: isr,
		})
	}
	return infos, nil
}

// FetchMessages fetches messages from a topic with global sorting.
// topic: topic name
// limit: number of messages to fetch
// sortOrder: 'oldest' or 'newest'
// Returns a slice of Message and error if fetching fails.
func (c *Client) FetchMessages(topic string, limit int, sortOrder string) ([]models.Message, error) {
	partitions, err := c.client.Partitions(topic)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumerFromClient(c.client)
	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	// For proper sorting, we need to fetch more messages than the limit
	// then sort globally and take the top N
	fetchMultiplier := 3 // Fetch 3x the limit to ensure we get the true top N
	if limit == 0 {
		fetchMultiplier = 1 // If no limit, don't multiply
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Channel to collect messages from all partitions
	messagesChan := make(chan models.Message, 1000) // Larger buffer
	var wg sync.WaitGroup

	// Calculate how many messages to fetch per partition
	messagesPerPartition := 50 // Default minimum per partition
	if limit > 0 {
		// Fetch more messages per partition to ensure we get the globally correct sorted order
		messagesPerPartition = max((limit*fetchMultiplier)/len(partitions), 20)
		messagesPerPartition = min(messagesPerPartition, 200) // Cap at 200 per partition
	}

	// Process partitions concurrently
	for _, partition := range partitions {
		wg.Add(1)
		go func(partitionID int32) {
			defer wg.Done()
			if sortOrder == "oldest" {
				c.fetchOldestFromPartition(ctx, consumer, topic, partitionID, messagesPerPartition, messagesChan)
			} else {
				c.fetchNewestFromPartition(ctx, consumer, topic, partitionID, messagesPerPartition, messagesChan)
			}
		}(partition)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(messagesChan)
	}()

	// Collect all messages
	var messages []models.Message
	for msg := range messagesChan {
		messages = append(messages, msg)
	}

	// Now sort all messages globally
	if sortOrder == "oldest" {
		sort.Slice(messages, func(i, j int) bool {
			if messages[i].Timestamp == messages[j].Timestamp {
				// Secondary sort: partition then offset for deterministic ordering
				if messages[i].Partition == messages[j].Partition {
					return messages[i].Offset < messages[j].Offset
				}
				return messages[i].Partition < messages[j].Partition
			}
			return messages[i].Timestamp < messages[j].Timestamp
		})
	} else {
		sort.Slice(messages, func(i, j int) bool {
			if messages[i].Timestamp == messages[j].Timestamp {
				// Secondary sort: partition then offset for deterministic ordering
				if messages[i].Partition == messages[j].Partition {
					return messages[i].Offset > messages[j].Offset
				}
				return messages[i].Partition > messages[j].Partition
			}
			return messages[i].Timestamp > messages[j].Timestamp
		})
	}

	// Apply final limit after global sorting
	if limit > 0 && len(messages) > limit {
		messages = messages[:limit]
	}

	return messages, nil
}

// fetchNewestFromPartition fetches the newest messages from a partition
func (c *Client) fetchNewestFromPartition(ctx context.Context, consumer sarama.Consumer, topic string, partition int32, limit int, messagesChan chan<- models.Message) {
	newest, err := c.client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return
	}
	oldest, err := c.client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return
	}

	if newest <= oldest {
		return
	}

	// For newest messages, start from the end and work backwards
	startOffset := newest - int64(limit)
	if startOffset < oldest {
		startOffset = oldest
	}

	pc, err := consumer.ConsumePartition(topic, partition, startOffset)
	if err != nil {
		return
	}
	defer pc.Close()

	var partitionMessages []models.Message
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	// Collect messages from this partition
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-pc.Messages():
			if msg == nil {
				break
			}

			msgHeaders := make([]models.MessageHeader, len(msg.Headers))
			for i, h := range msg.Headers {
				msgHeaders[i] = models.MessageHeader{
					Key:   string(h.Key),
					Value: string(h.Value),
				}
			}

			message := models.Message{
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       string(msg.Key),
				Value:     string(msg.Value),
				Timestamp: msg.Timestamp.UnixMilli(),
				Size:      len(msg.Value) + len(msg.Key),
				Headers:   msgHeaders,
			}

			partitionMessages = append(partitionMessages, message)

			// Reset timeout since we got a message
			if !timeout.Stop() {
				<-timeout.C
			}
			timeout.Reset(3 * time.Second)

		case <-timeout.C:
			// Timeout reached, send what we have
			goto sendMessages
		case err := <-pc.Errors():
			if err != nil {
				fmt.Printf("Error consuming from partition %d: %v\n", partition, err)
			}
			goto sendMessages
		}
	}

sendMessages:
	// Send all messages from this partition to the channel
	for _, msg := range partitionMessages {
		select {
		case messagesChan <- msg:
		case <-ctx.Done():
			return
		}
	}
}

// fetchOldestFromPartition fetches the oldest messages from a partition
func (c *Client) fetchOldestFromPartition(ctx context.Context, consumer sarama.Consumer, topic string, partition int32, limit int, messagesChan chan<- models.Message) {
	newest, err := c.client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return
	}
	oldest, err := c.client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return
	}

	if newest <= oldest {
		return
	}

	// For oldest messages, start from the beginning
	pc, err := consumer.ConsumePartition(topic, partition, oldest)
	if err != nil {
		return
	}
	defer pc.Close()

	messageCount := 0
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-pc.Messages():
			if msg == nil {
				return
			}

			msgHeaders := make([]models.MessageHeader, len(msg.Headers))
			for i, h := range msg.Headers {
				msgHeaders[i] = models.MessageHeader{
					Key:   string(h.Key),
					Value: string(h.Value),
				}
			}

			message := models.Message{
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       string(msg.Key),
				Value:     string(msg.Value),
				Timestamp: msg.Timestamp.UnixMilli(),
				Size:      len(msg.Value) + len(msg.Key),
				Headers:   msgHeaders,
			}

			select {
			case messagesChan <- message:
				messageCount++
				// Reset timeout since we got a message
				if !timeout.Stop() {
					<-timeout.C
				}
				timeout.Reset(3 * time.Second)

				// Stop if we've reached the limit for this partition
				if messageCount >= limit {
					return
				}
			case <-ctx.Done():
				return
			}

		case <-timeout.C:
			return
		case err := <-pc.Errors():
			if err != nil {
				fmt.Printf("Error consuming from partition %d: %v\n", partition, err)
			}
			return
		}
	}
}

// FetchRecentMessages - optimized method for getting recent messages quickly
func (c *Client) FetchRecentMessages(topic string, limit int) ([]models.Message, error) {
	return c.FetchMessages(topic, limit, "newest")
}

// FetchAllMessages - method to get all available messages (use with caution)
func (c *Client) FetchAllMessages(topic string, sortOrder string) ([]models.Message, error) {
	return c.FetchMessages(topic, 0, sortOrder)
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Produce produces a message to a topic
func (c *Client) Produce(topic, key string, value []byte, partition int32, headers []models.MessageHeader) error {
	// Defensive: check if requested partition exists
	if partition >= 0 {
		partitions, err := c.client.Partitions(topic)
		if err != nil {
			return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
		}
		valid := false
		for _, p := range partitions {
			if p == partition {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("partition %d does not exist for topic %s", partition, topic)
		}
	}

	producer, err := sarama.NewSyncProducerFromClient(c.client)
	if err != nil {
		return err
	}
	defer producer.Close()

	var saramaHeaders []sarama.RecordHeader
	for _, h := range headers {
		saramaHeaders = append(saramaHeaders, sarama.RecordHeader{
			Key:   []byte(h.Key),
			Value: []byte(h.Value),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: saramaHeaders,
	}

	// Set partition - this will only work with manual partitioner
	if partition >= 0 {
		msg.Partition = partition
	}
	// If partition is -1 or not specified, let the partitioner decide

	actualPartition, offset, err := producer.SendMessage(msg)
	if err != nil {
		// Log error for diagnostics
		fmt.Println("Produce error:", err.Error())
		fmt.Printf("Topic: %s, Key: %s, Partition (requested): %d, ValueLen: %d\n", topic, key, partition, len(value))
		return errors.New("produce failed: " + err.Error())
	}

	fmt.Printf("Produce success: Topic=%s, Key=%s, Partition (requested)=%d, Partition (actual)=%d, Offset=%d, ValueLen=%d\n",
		topic, key, partition, actualPartition, offset, len(value))

	return nil
}

// Alternative: Create a producer with specific partitioner configuration
func (c *Client) ProduceWithCustomPartitioner(topic, key string, value []byte, partition int32, headers []models.MessageHeader) error {
	// Create a custom config for this producer with manual partitioner
	producerConfig := *c.config // Copy the config
	producerConfig.Producer.Partitioner = sarama.NewManualPartitioner

	producer, err := sarama.NewSyncProducer(c.brokers, &producerConfig)
	if err != nil {
		return err
	}
	defer producer.Close()

	var saramaHeaders []sarama.RecordHeader
	for _, h := range headers {
		saramaHeaders = append(saramaHeaders, sarama.RecordHeader{
			Key:   []byte(h.Key),
			Value: []byte(h.Value),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(value),
		Headers:   saramaHeaders,
		Partition: partition, // This will be respected with manual partitioner
	}

	actualPartition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("produce failed: %w", err)
	}

	fmt.Printf("Produce success: Topic=%s, Partition (requested)=%d, Partition (actual)=%d, Offset=%d\n",
		topic, partition, actualPartition, offset)

	return nil
}

// ClearTopicMessages clears all messages from a topic by truncating to latest offset
func (c *Client) ClearTopicMessages(topic string) error {
	// Get all partitions for the topic
	partitions, err := c.client.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions for topic %s: %w", topic, err)
	}

	// For each partition, set the offset to the latest (effectively truncating all messages)
	for _, partition := range partitions {
		latest, err := c.client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to get latest offset for partition %d: %w", partition, err)
		}

		// Create a record to truncate to latest offset
		request := &sarama.DeleteRecordsRequest{
			Version: 1,
			Topics: map[string]*sarama.DeleteRecordsRequestTopic{
				topic: {
					PartitionOffsets: map[int32]int64{
						partition: latest,
					},
				},
			},
		}

		// Get any broker to send the request
		brokers := c.client.Brokers()
		if len(brokers) == 0 {
			return errors.New("no brokers available")
		}

		_, err = brokers[0].DeleteRecords(request)
		if err != nil {
			return fmt.Errorf("failed to delete records for partition %d: %w", partition, err)
		}
	}

	fmt.Printf("Successfully cleared all messages from topic: %s\n", topic)
	return nil
}

// ClearTopicMessagesWithRetention clears messages by temporarily setting retention to 1ms
func (c *Client) ClearTopicMessagesWithRetention(topic string) error {
	// First, get current topic configuration
	configEntries, err := c.admin.DescribeConfig(sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topic,
	})
	if err != nil {
		return fmt.Errorf("failed to describe topic config: %w", err)
	}

	// Store original retention settings
	var originalRetentionMs, originalRetentionBytes string
	for _, entry := range configEntries {
		if entry.Name == "retention.ms" {
			originalRetentionMs = entry.Value
		}
		if entry.Name == "retention.bytes" {
			originalRetentionBytes = entry.Value
		}
	}

	// Set retention to 1ms to force immediate cleanup
	alterConfig := map[string]*string{
		"retention.ms": stringPtr("1"),
	}

	if originalRetentionBytes != "" {
		alterConfig["retention.bytes"] = stringPtr("1")
	}

	err = c.admin.AlterConfig(sarama.TopicResource, topic, alterConfig, false)
	if err != nil {
		return fmt.Errorf("failed to set temporary retention: %w", err)
	}

	fmt.Printf("Set temporary retention for topic %s, waiting for cleanup...\n", topic)

	// Wait for cleanup to happen (this depends on log.segment.ms and log.retention.check.interval.ms)
	time.Sleep(10 * time.Second)

	// Restore original retention settings
	restoreConfig := map[string]*string{}
	if originalRetentionMs != "" {
		restoreConfig["retention.ms"] = &originalRetentionMs
	} else {
		// Reset to default (7 days)
		restoreConfig["retention.ms"] = stringPtr("604800000")
	}

	if originalRetentionBytes != "" {
		restoreConfig["retention.bytes"] = &originalRetentionBytes
	}

	err = c.admin.AlterConfig(sarama.TopicResource, topic, restoreConfig, false)
	if err != nil {
		return fmt.Errorf("failed to restore original retention: %w", err)
	}

	fmt.Printf("Successfully cleared messages and restored retention for topic: %s\n", topic)
	return nil
}

// GetBrokers returns broker information
func (c *Client) GetBrokers() ([]models.Broker, error) {
	brokers := c.client.Brokers()
	topicNames, err := c.client.Topics()
	if err != nil {
		return nil, err
	}
	details, err := c.admin.DescribeTopics(topicNames)
	if err != nil {
		return nil, err
	}
	// Build maps of brokerID to leaders and replicas
	leaderMap := make(map[int32][]int)
	replicaMap := make(map[int32][]int)
	segmentCountMap := make(map[int32]int)
	for _, topic := range details {
		for _, part := range topic.Partitions {
			leader := part.Leader
			leaderMap[leader] = append(leaderMap[leader], int(part.ID))
			segmentCountMap[leader]++
			for _, r := range part.Replicas {
				replicaMap[r] = append(replicaMap[r], int(part.ID))
			}
		}
	}
	var infos []models.Broker
	for _, b := range brokers {
		addr := b.Addr()
		id := b.ID()
		port := 0
		host := addr
		if h, p, err := net.SplitHostPort(addr); err == nil {
			host = h
			if parsed, err := strconv.Atoi(p); err == nil {
				port = parsed
			}
		}
		infos = append(infos, models.Broker{
			ID:           id,
			Host:         host,
			Port:         int32(port),
			Address:      addr,
			Status:       "online",
			SegmentCount: segmentCountMap[id],
			Replicas:     replicaMap[id],
			Leaders:      leaderMap[id],
		})
	}
	return infos, nil
}

// GetConsumers returns consumer group information
func (c *Client) GetConsumers() ([]models.ConsumerGroup, error) {
	if c.admin == nil {
		return nil, fmt.Errorf("sarama admin client not initialized")
	}
	groups, err := c.admin.ListConsumerGroups()
	if err != nil {
		return nil, err
	}
	var infos []models.ConsumerGroup
	for groupID := range groups {
		desc, err := c.admin.DescribeConsumerGroups([]string{groupID})
		if err != nil || len(desc) == 0 {
			continue
		}
		cg := desc[0]
		for _, member := range cg.Members {
			assignment, err := member.GetMemberAssignment()
			if err != nil {
				infos = append(infos, models.ConsumerGroup{
					GroupID:    groupID,
					MemberID:   member.MemberId,
					Topics:     []string{},
					Partitions: []int32{},
					Error:      err.Error(),
				})
				continue
			}
			var topics []string
			var partitions []int32
			for topic, parts := range assignment.Topics {
				topics = append(topics, topic)
				partitions = append(partitions, parts...)
			}
			infos = append(infos, models.ConsumerGroup{
				GroupID:    groupID,
				MemberID:   member.MemberId,
				Topics:     topics,
				Partitions: partitions,
				Error:      "",
			})
		}
	}
	return infos, nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// convertReplicas converts Kafka broker IDs to integers
func convertReplicas(replicas []int32) []int {
	result := make([]int, len(replicas))
	for i, r := range replicas {
		result[i] = int(r)
	}
	return result
}

// DeleteTopic deletes a topic using the Sarama admin client
func (c *Client) DeleteTopic(topic string) error {
	if c.admin == nil {
		return fmt.Errorf("admin client not initialized")
	}
	return c.admin.DeleteTopic(topic)
}
