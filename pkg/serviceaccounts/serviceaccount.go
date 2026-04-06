package serviceaccounts

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// ErrServiceAccountNotFound is returned when a service account lookup fails.
var ErrServiceAccountNotFound = errors.New("service account not found")

// ErrServiceAccountExists is returned when creating a duplicate service account.
var ErrServiceAccountExists = errors.New("service account already exists")

// StringList is a []string that serialises as a comma-separated value in SQL
// and as a JSON array over the wire.
//
// TODO: revisit when org-scoping is added — AllowedServices may need to become
// a join table keyed by (service_account_id, org_id, service_id).
type StringList []string

func (s StringList) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New("unsupported scan type for StringList")
	}
	if str == "" {
		*s = StringList{}
		return nil
	}
	*s = strings.Split(str, ",")
	return nil
}

// ServiceAccount is a machine identity that can request virtual keys without
// holding full management-level user privileges.
//
// NOTE (flat v1): service accounts are global to the Bifrost instance.
// Org-scoped service accounts are deferred — see docs/backlog.md.
// AllowedServices being empty means the account may request a key for any
// registered service.
type ServiceAccount struct {
	ID              string     `json:"id" gorm:"primaryKey;size:255"`
	Name            string     `json:"name" gorm:"not null;size:255"`
	APIKey          string     `json:"api_key" gorm:"not null;uniqueIndex;size:255"`
	AllowedServices StringList `json:"allowed_services" gorm:"type:text"`
}

func (ServiceAccount) TableName() string { return "service_accounts" }
