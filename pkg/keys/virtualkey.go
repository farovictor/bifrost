package keys

import "time"

// VirtualKey represents a short-lived key granting access to a target service.
// The JSON fields use snake_case to match API payloads.
type VirtualKey struct {
	ID        string    `json:"id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	Target    string    `json:"target"`
}
