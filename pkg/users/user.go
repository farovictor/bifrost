package users

// User represents an API user able to authenticate to Bifrost.
type User struct {
	ID     string `json:"id"`
	APIKey string `json:"api_key"`
}
