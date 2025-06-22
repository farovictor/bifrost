package users

// User represents an API user able to authenticate to Bifrost.
type User struct {
	ID     string `json:"id" gorm:"primaryKey;size:255"`
	APIKey string `json:"api_key" gorm:"not null"`
}

func (User) TableName() string { return "users" }
