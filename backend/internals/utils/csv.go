package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	userMutex sync.RWMutex
)

// csv.go - Provides utilities for managing user data stored in a CSV file.
// Includes functions for retrieving and updating users, and ensuring the users file exists.

// User represents a user in the system, with a username and password.
type User struct {
	Username string // Username of the user
	Password string // Password of the user (plain text; consider hashing in production)
}

// ensureUsersFile creates the users.csv file if it doesn't exist, and adds a default admin user if needed.
func ensureUsersFile() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	filePath := filepath.Join("data", UsersFileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create file with header
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create users file: %v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		if err := writer.Write([]string{"username", "password"}); err != nil {
			return fmt.Errorf("failed to write header: %v", err)
		}

		// Write default admin user
		if err := writer.Write([]string{"admin", "password"}); err != nil {
			return fmt.Errorf("failed to write default user: %v", err)
		}
	}
	return nil
}

// GetUser retrieves a user by username from the CSV file.
// Returns a pointer to User and error if not found or on failure.
func GetUser(username string) (*User, error) {
	userMutex.RLock()
	defer userMutex.RUnlock()

	if err := ensureUsersFile(); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join("data", UsersFileName), os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open users file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read users file: %v", err)
	}

	// Skip header row
	for _, record := range records[1:] {
		if record[0] == username {
			return &User{
				Username: record[0],
				Password: record[1],
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// UpdateUserPassword updates a user's password in the CSV file.
// Returns error if the user is not found or on failure.
func UpdateUserPassword(username, newPassword string) error {
	userMutex.Lock()
	defer userMutex.Unlock()

	if err := ensureUsersFile(); err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join("data", UsersFileName), os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open users file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read users file: %v", err)
	}

	// Find and update the user's password
	userFound := false
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		if record[0] == username {
			records[i][1] = newPassword
			userFound = true
			break
		}
	}

	if !userFound {
		return fmt.Errorf("user not found")
	}

	// Write back to file
	file, err = os.Create(filepath.Join("data", UsersFileName))
	if err != nil {
		return fmt.Errorf("failed to create users file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write users file: %v", err)
	}

	return nil
}
