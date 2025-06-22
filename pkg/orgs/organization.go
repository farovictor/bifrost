package orgs

// Organization represents a collection of resources under a single tenant.
type Organization struct {
	ID     string `json:"id" gorm:"primaryKey;size:255;default:uuid_generate_v4()"`
	Name   string `json:"name" gorm:"not null"`
	Domain string `json:"domain" gorm:"not null"`
	Email  string `json:"email" gorm:"not null"`
}

func (Organization) TableName() string { return "organizations" }
