package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"backend/internals/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// auth.go - Provides JWT authentication middleware for protecting API routes.
// Validates JWT tokens, extracts user claims, and attaches them to the request context.

// JWTMiddleware validates JWT tokens and adds claims to the request context.
// Use this middleware to protect routes that require authentication.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or malformed token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract claims and attach user info to context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok && time.Now().Unix() > int64(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				return
			}
			// Store user info (e.g., username or user ID) in context for downstream handlers
			c.Set("user", claims["sub"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
}

// Use env var for JWT secret, fallback to constant
func getJWTSecret() []byte {
	secret := os.Getenv(utils.JWTSecretKeyEnv)
	if secret == "" {
		secret = utils.DefaultJWTSecret
	}
	return []byte(secret)
}
