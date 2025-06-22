package rootkeys

// RootKey represents a real API key stored separately from services.
type RootKey struct {
	ID     string `json:"id" gorm:"primaryKey;size:255"`
	APIKey string `json:"api_key" gorm:"not null"`
}

func (RootKey) TableName() string { return "root_keys" }
