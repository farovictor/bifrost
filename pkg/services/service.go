package services

// CredentialStyle constants for how the root key is injected into upstream requests.
const (
	// CredentialHeaderXAPIKey injects the key as X-API-Key (default).
	CredentialHeaderXAPIKey = "X-API-Key"
	// CredentialHeaderBearer injects the key as Authorization: Bearer <key>.
	CredentialHeaderBearer = "Authorization"
)

// Service represents an upstream service to which requests can be proxied.
type Service struct {
	ID               string `json:"id" gorm:"primaryKey;size:255"`
	Endpoint         string `json:"endpoint" gorm:"not null"`
	RootKeyID        string `json:"root_key_id" gorm:"not null"`
	CredentialHeader string `json:"credential_header,omitempty" gorm:"size:255"`
}

func (Service) TableName() string { return "services" }
