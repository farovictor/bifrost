package orgs

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// Store defines persistence behavior for Organization objects.
type Store interface {
	Create(Organization) error
	Get(id string) (Organization, error)
	Delete(id string) error
	Update(Organization) error
	List() []Organization
}

// MemoryStore keeps Organizations in memory with concurrency safety.
type MemoryStore struct {
	mu   sync.RWMutex
	orgs map[string]Organization
}

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{orgs: make(map[string]Organization)}
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *gorm.DB) *PostgresStore {
	db.AutoMigrate(&Organization{})
	return &PostgresStore{db: db}
}

// PostgresStore persists organizations in PostgreSQL.
type PostgresStore struct {
	db *gorm.DB
}

// Create inserts a new Organization. Returns error if ID already exists.
func (s *MemoryStore) Create(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o.ID == "" {
		o.ID = GenerateID()
	}
	if _, ok := s.orgs[o.ID]; ok {
		return ErrOrgExists
	}
	s.orgs[o.ID] = o
	return nil
}

// Get retrieves an Organization by ID.
func (s *MemoryStore) Get(id string) (Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orgs[id]
	if !ok {
		return Organization{}, ErrOrgNotFound
	}
	return o, nil
}

// Delete removes an Organization from the store.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[id]; !ok {
		return ErrOrgNotFound
	}
	delete(s.orgs, id)
	return nil
}

// Update replaces an existing Organization.
func (s *MemoryStore) Update(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[o.ID]; !ok {
		return ErrOrgNotFound
	}
	s.orgs[o.ID] = o
	return nil
}

// List returns all Organizations currently in the store.
func (s *MemoryStore) List() []Organization {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Organization, 0, len(s.orgs))
	for _, o := range s.orgs {
		out = append(out, o)
	}
	return out
}

// Create inserts an organization into the database.
func (s *PostgresStore) Create(o Organization) error {
	if o.ID == "" {
		o.ID = GenerateID()
	}
	if err := s.db.Create(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrOrgExists
		}
		return err
	}
	return nil
}

// Get retrieves an organization by ID.
func (s *PostgresStore) Get(id string) (Organization, error) {
	var o Organization
	if err := s.db.First(&o, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Organization{}, ErrOrgNotFound
		}
		return Organization{}, err
	}
	return o, nil
}

// Delete removes an organization.
func (s *PostgresStore) Delete(id string) error {
	res := s.db.Delete(&Organization{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrOrgNotFound
	}
	return nil
}

// Update replaces an organization.
func (s *PostgresStore) Update(o Organization) error {
	res := s.db.Model(&Organization{}).Where("id = ?", o.ID).Updates(o)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrOrgNotFound
	}
	return nil
}

// List returns all organizations.
func (s *PostgresStore) List() []Organization {
	var out []Organization
	if err := s.db.Find(&out).Error; err != nil {
		return nil
	}
	return out
}

var (
	ErrOrgNotFound = errors.New("organization not found")
	ErrOrgExists   = errors.New("organization already exists")
)
