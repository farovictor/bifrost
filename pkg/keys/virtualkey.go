package keys

import "time"

// VirtualKey represents a short-lived key granting access to a target service.
// The JSON fields use snake_case to match API payloads.
type VirtualKey struct {
	ID        string    `json:"id" gorm:"primaryKey;size:255"`
	Scope     string    `json:"scope" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Target    string    `json:"target" gorm:"not null"`
	RateLimit int       `json:"rate_limit" gorm:"not null"`
	Source    string    `json:"source,omitempty" gorm:"size:16;default:''"`
	OneShot   bool      `json:"one_shot,omitempty" gorm:"default:false"`
	Used      bool      `json:"used,omitempty" gorm:"default:false"`
}

// SourceMCP is the source label for keys issued via the MCP tool.
const SourceMCP = "mcp"

func (VirtualKey) TableName() string { return "virtual_keys" }
