package usage

import "time"

// Event records a single proxied request for a virtual key.
type Event struct {
	ID               uint      `json:"id"                gorm:"primaryKey;autoIncrement"`
	KeyID            string    `json:"key_id"            gorm:"not null;index"`
	Timestamp        time.Time `json:"timestamp"         gorm:"not null;index"`
	StatusCode       int       `json:"status_code"       gorm:"not null"`
	Service          string    `json:"service"           gorm:"not null;size:255"`
	LatencyMS        int64     `json:"latency_ms"        gorm:"not null"`
	PromptTokens     int       `json:"prompt_tokens,omitempty"`
	CompletionTokens int       `json:"completion_tokens,omitempty"`
	TotalTokens      int       `json:"total_tokens,omitempty"`
}

func (Event) TableName() string { return "usage_events" }
