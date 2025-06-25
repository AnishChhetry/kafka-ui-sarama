package utils

import "errors"

// errors.go - Centralized error variables for the backend application.

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrUserNotAuthenticated     = errors.New("user not authenticated")
	ErrUserNotFound             = errors.New("user not found")
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
	ErrFailedToUpdatePassword   = errors.New("failed to update password")
	ErrInvalidRequestBody       = errors.New("invalid request body")
	ErrTokenExpired             = errors.New("token expired")
	ErrInvalidToken             = errors.New("invalid token")
	ErrMissingToken             = errors.New("missing or malformed token")
	ErrInvalidTokenClaims       = errors.New("invalid token claims")
	ErrBootstrapServerRequired  = errors.New("bootstrapServer parameter is required")
	ErrTopicCreationFailed      = errors.New("failed to create topic")
	ErrMessageProductionFailed  = errors.New("failed to produce message")
)
