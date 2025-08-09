package orgs

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// MembershipStore defines persistence behavior for Membership objects.
type MembershipStore interface {
	Create(Membership) error
	Get(userID, orgID string) (Membership, error)
	Delete(userID, orgID string) error
	Update(Membership) error
	List() []Membership
	ListByUser(userID string) []Membership
}

// SQLMembershipStore persists memberships in a SQL database and implements MembershipStore.
// It mirrors the in-memory MemoryMembershipStore behavior.
type SQLMembershipStore struct {
	db *gorm.DB
}

// NewSQLMembershipStore creates a SQL-backed store and auto-migrates the Membership model.
func NewSQLMembershipStore(db *gorm.DB) *SQLMembershipStore {
	db.AutoMigrate(&Membership{})
	return &SQLMembershipStore{db: db}
}

// Create inserts a new membership. Returns error if the pair already exists.
func (s *SQLMembershipStore) Create(m Membership) error {
	if err := s.db.Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrMembershipExists
		}
		return err
	}
	return nil
}

// Get retrieves a membership by user and organization IDs.
func (s *SQLMembershipStore) Get(userID, orgID string) (Membership, error) {
	var m Membership
	if err := s.db.First(&m, "user_id = ? AND org_id = ?", userID, orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Membership{}, ErrMembershipNotFound
		}
		return Membership{}, err
	}
	return m, nil
}

// Delete removes a membership.
func (s *SQLMembershipStore) Delete(userID, orgID string) error {
	res := s.db.Delete(&Membership{}, "user_id = ? AND org_id = ?", userID, orgID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMembershipNotFound
	}
	return nil
}

// Update replaces an existing membership.
func (s *SQLMembershipStore) Update(m Membership) error {
	res := s.db.Model(&Membership{}).Where("user_id = ? AND org_id = ?", m.UserID, m.OrgID).Updates(m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMembershipNotFound
	}
	return nil
}

// List returns all memberships.
func (s *SQLMembershipStore) List() []Membership {
	var out []Membership
	if err := s.db.Find(&out).Error; err != nil {
		return nil
	}
	return out
}

// ListByUser returns memberships for a specific user.
func (s *SQLMembershipStore) ListByUser(userID string) []Membership {
	var out []Membership
	if err := s.db.Where("user_id = ?", userID).Find(&out).Error; err != nil {
		return nil
	}
	return out
}

var (
	ErrMembershipNotFound = errors.New("membership not found")
	ErrMembershipExists   = errors.New("membership already exists")
)
