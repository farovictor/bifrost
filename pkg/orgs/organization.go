package orgs

// Organization represents a collection of resources under a single tenant.
type Organization struct {
	ID   string `json:"id" gorm:"primaryKey;size:255"`
	Name string `json:"name" gorm:"not null"`
}

func (Organization) TableName() string { return "organizations" }
