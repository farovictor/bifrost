package services

// Service represents an upstream service to which requests can be proxied.
// Service represents an upstream service to which requests can be proxied.
type Service struct {
	ID        string `json:"id"`
	Endpoint  string `json:"endpoint"`
	RootKeyID string `json:"root_key_id"`
}
