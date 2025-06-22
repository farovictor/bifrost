package services

// Service represents an upstream service to which requests can be proxied.
type Service struct {
	ID        string `json:"id" gorm:"primaryKey;size:255"`
	Endpoint  string `json:"endpoint" gorm:"not null"`
	RootKeyID string `json:"root_key_id" gorm:"not null"`
}

func (Service) TableName() string { return "services" }
