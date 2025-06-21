package rootkeys

// RootKey represents a real API key stored separately from services.
type RootKey struct {
	ID     string `json:"id"`
	APIKey string `json:"api_key"`
}
