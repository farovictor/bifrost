package orgs

import (
	"errors"

	"gorm.io/gorm"
)

// PostgresMembershipStore persists memberships in PostgreSQL.
// It mirrors the in-memory MemoryMembershipStore behavior.
type PostgresMembershipStore struct {
	db *gorm.DB
}

// NewPostgresMembershipStore creates a Postgres-backed store.
func NewPostgresMembershipStore(db *gorm.DB) *PostgresMembershipStore {
	db.AutoMigrate(&Membership{})
	return &PostgresMembershipStore{db: db}
}

// Create inserts a new membership. Returns error if the pair already exists.
func (s *PostgresMembershipStore) Create(m Membership) error {
	if err := s.db.Create(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrMembershipExists
		}
		return err
	}
	return nil
}

// Get retrieves a membership by user and organization IDs.
func (s *PostgresMembershipStore) Get(userID, orgID string) (Membership, error) {
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
func (s *PostgresMembershipStore) Delete(userID, orgID string) error {
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
func (s *PostgresMembershipStore) Update(m Membership) error {
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
func (s *PostgresMembershipStore) List() []Membership {
	var out []Membership
	if err := s.db.Find(&out).Error; err != nil {
		return nil
	}
	return out
}

// ListByUser returns memberships for a specific user.
func (s *PostgresMembershipStore) ListByUser(userID string) []Membership {
	var out []Membership
	if err := s.db.Where("user_id = ?", userID).Find(&out).Error; err != nil {
		return nil
	}
	return out
}
