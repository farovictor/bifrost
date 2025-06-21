package services

// Service represents an upstream service to which requests can be proxied.
type Service struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
}
