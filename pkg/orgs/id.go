package orgs

import "github.com/google/uuid"

// GenerateID returns a new UUIDv4 string.
func GenerateID() string {
	return uuid.NewString()
}
