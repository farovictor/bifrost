package keys

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// Store defines persistence behavior for VirtualKey objects.
type Store interface {
	Create(VirtualKey) error
	Get(id string) (VirtualKey, error)
	Update(id string, k VirtualKey) error
	Delete(id string) error
	List() []VirtualKey
}

// MemoryStore is an in-memory repository for VirtualKey objects.
type MemoryStore struct {
	mu   sync.RWMutex
	keys map[string]VirtualKey
}

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]VirtualKey)}
}

// NewSQLStore creates a SQL-backed store.
func NewSQLStore(db *gorm.DB) *SQLStore {
	db.AutoMigrate(&VirtualKey{})
	return &SQLStore{db: db}
}

// SQLStore persists VirtualKeys in a SQL database.
type SQLStore struct {
	db *gorm.DB
}

// Create inserts a new VirtualKey. Returns an error if the key ID already exists.
func (s *MemoryStore) Create(k VirtualKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

// Get retrieves a VirtualKey by its ID.
func (s *MemoryStore) Get(id string) (VirtualKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.keys[id]
	if !ok {
		return VirtualKey{}, ErrKeyNotFound
	}
	return v, nil
}

// Update replaces the VirtualKey stored under the given ID.
func (s *MemoryStore) Update(id string, k VirtualKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	s.keys[id] = k
	return nil
}

// Delete removes a VirtualKey from the store.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

// List returns all VirtualKeys currently in the store.
func (s *MemoryStore) List() []VirtualKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VirtualKey, 0, len(s.keys))
	for _, v := range s.keys {
		out = append(out, v)
	}
	return out
}

// Create inserts a virtual key into the database.
func (s *SQLStore) Create(k VirtualKey) error {
	if err := s.db.Create(&k).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrKeyExists
		}
		return err
	}
	return nil
}

// Get retrieves a virtual key by ID.
func (s *SQLStore) Get(id string) (VirtualKey, error) {
	var v VirtualKey
	if err := s.db.First(&v, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return VirtualKey{}, ErrKeyNotFound
		}
		return VirtualKey{}, err
	}
	return v, nil
}

// Update replaces an existing virtual key.
func (s *SQLStore) Update(id string, k VirtualKey) error {
	res := s.db.Model(&VirtualKey{}).Where("id = ?", id).Updates(k)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Delete removes a virtual key by ID.
func (s *SQLStore) Delete(id string) error {
	res := s.db.Delete(&VirtualKey{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// List returns all virtual keys from the database.
func (s *SQLStore) List() []VirtualKey {
	var out []VirtualKey
	if err := s.db.Find(&out).Error; err != nil {
		return nil
	}
	return out
}

// Error values returned by Store operations.
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)
