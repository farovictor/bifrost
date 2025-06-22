// Package utils provides helper functions used across the project.
package utils

import "github.com/google/uuid"

// GenerateID returns a new UUIDv4 string.
func GenerateID() string {
	return uuid.NewString()
}
