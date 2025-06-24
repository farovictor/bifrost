package orgs

import (
	"errors"
	"sync"

	"github.com/farovictor/bifrost/pkg/utils"
	"gorm.io/gorm"
)

// Store defines persistence behavior for Organization objects.
type Store interface {
	Create(Organization) error
	Get(id string) (Organization, error)
	GetByName(name string) (Organization, error)
	Delete(id string) error
	Update(Organization) error
	List() []Organization
}

// MemoryStore keeps Organizations in memory with concurrency safety.
type MemoryStore struct {
	mu    sync.RWMutex
	orgs  map[string]Organization
	names map[string]string
}

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{orgs: make(map[string]Organization), names: make(map[string]string)}
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
		o.ID = utils.GenerateID()
	}
	if _, ok := s.names[o.Name]; ok {
		return ErrOrgExists
	}
	if _, ok := s.orgs[o.ID]; ok {
		return ErrOrgExists
	}
	s.orgs[o.ID] = o
	s.names[o.Name] = o.ID
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

// GetByName retrieves an Organization by name.
func (s *MemoryStore) GetByName(name string) (Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.names[name]
	if !ok {
		return Organization{}, ErrOrgNotFound
	}
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
	o, ok := s.orgs[id]
	if !ok {
		return ErrOrgNotFound
	}
	delete(s.orgs, id)
	delete(s.names, o.Name)
	return nil
}

// Update replaces an existing Organization.
func (s *MemoryStore) Update(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	curr, ok := s.orgs[o.ID]
	if !ok {
		return ErrOrgNotFound
	}
	if curr.Name != o.Name {
		if _, exists := s.names[o.Name]; exists {
			return ErrOrgExists
		}
		delete(s.names, curr.Name)
		s.names[o.Name] = o.ID
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
		o.ID = utils.GenerateID()
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

// GetByName retrieves an organization by name.
func (s *PostgresStore) GetByName(name string) (Organization, error) {
	var o Organization
	if err := s.db.First(&o, "name = ?", name).Error; err != nil {
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
