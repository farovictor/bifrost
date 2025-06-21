package orgs

// Organization represents a collection of resources under a single tenant.
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
