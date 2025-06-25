package middleware

import (
	"backend/internals/kafka"

	"github.com/gin-gonic/gin"
)

var kafkaService kafka.KafkaService

// bootstrap.go - Provides middleware for handling Kafka bootstrap server configuration.
// Allows dynamic updating of the Kafka client based on the bootstrap server provided in requests.

// SetKafkaService sets the Kafka service instance for use by middleware and handlers.
func SetKafkaService(service kafka.KafkaService) {
	kafkaService = service
}

// BootstrapMiddleware updates the Kafka client if a new bootstrap server is provided in the request.
// Use this middleware to allow clients to dynamically set the Kafka broker address.
func BootstrapMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bootstrapServer := c.Query("bootstrapServer")
		if bootstrapServer != "" {
			// Create a new Kafka client with the provided broker address
			brokers := []string{bootstrapServer}
			newClient, err := kafka.NewKafkaClient(brokers, nil)
			if err == nil {
				kafkaService = newClient
			}
		}
		c.Next()
	}
}
