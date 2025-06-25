package api

import (
	"net/http"
	"os"
	"time"

	"backend/internals/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// auth.go - Handles authentication-related API endpoints for login and password management.
// Provides JWT-based authentication and password change functionality.
//
// Endpoints:
//   - POST /login: Authenticate user and return JWT token
//   - POST /change-password: Change user password (requires authentication)
//
// Author: [Your Name]
// Date: [Date]

// Use env var for JWT secret, fallback to constant
func getJWTSecret() []byte {
	secret := os.Getenv(utils.JWTSecretKeyEnv)
	if secret == "" {
		secret = utils.DefaultJWTSecret
	}
	return []byte(secret)
}

// Login handles user authentication. It validates credentials and returns a JWT token if successful.
//
// Request JSON body:
//
//	{
//	  "username": "<username>",
//	  "password": "<password>"
//	}
//
// Response:
//
//	200 OK: { "token": "<jwt_token>" }
//	400 Bad Request: { "error": "Invalid request body" }
//	401 Unauthorized: { "error": "Invalid credentials" }
func Login(c *gin.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Retrieve user from persistent storage (CSV file)
	user, err := utils.GetUser(creds.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare provided password with stored password
	if user.Password != creds.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create JWT token with username as subject and 24h expiration
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": creds.Username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// ChangePassword allows an authenticated user to change their password.
//
// Request JSON body:
//
//	{
//	  "currentPassword": "<current_password>",
//	  "newPassword": "<new_password>"
//	}
//
// Response:
//
//	200 OK: { "message": "Password changed successfully" }
//	400 Bad Request: { "error": "Invalid request body" }
//	401 Unauthorized: { "error": "User not authenticated" / "Current password is incorrect" }
//	500 Internal Server Error: { "error": "Failed to update password" }
func ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Extract username from JWT token (set by authentication middleware)
	username, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Retrieve user from persistent storage (CSV file)
	user, err := utils.GetUser(username.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Verify current password matches stored password
	if user.Password != req.CurrentPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Update password in persistent storage
	if err := utils.UpdateUserPassword(username.(string), req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
