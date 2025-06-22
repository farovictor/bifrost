package orgs

type Membership struct {
	UserID string `json:"user_id" gorm:"primaryKey;size:255"`
	OrgID  string `json:"org_id" gorm:"primaryKey;size:255"`
	Role   string `json:"role" gorm:"not null"`
}

func (Membership) TableName() string { return "org_memberships" }
