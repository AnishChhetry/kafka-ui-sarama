package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// response.go - Provides utility functions for sending standardized JSON responses in HTTP handlers.

// JSONSuccess sends a 200 OK response with a success status and payload.
func JSONSuccess(c *gin.Context, payload interface{}) {
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": payload})
}

// JSONError sends a 500 Internal Server Error response with an error message.
func JSONError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
}
