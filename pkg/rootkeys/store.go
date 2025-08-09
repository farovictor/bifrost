package rootkeys

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// Store defines persistence behavior for RootKey objects.
type Store interface {
	Create(RootKey) error
	Get(id string) (RootKey, error)
	Delete(id string) error
	Update(RootKey) error
}

// MemoryStore keeps RootKeys in memory with concurrency safety.
type MemoryStore struct {
	mu   sync.RWMutex
	keys map[string]RootKey
}

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]RootKey)}
}

// NewSQLStore creates a SQL-backed store.
func NewSQLStore(db *gorm.DB) *SQLStore {
	db.AutoMigrate(&RootKey{})
	return &SQLStore{db: db}
}

// SQLStore persists RootKeys in a SQL database.
type SQLStore struct {
	db *gorm.DB
}

// Create inserts a new RootKey. Returns error if ID already exists.
func (s *MemoryStore) Create(k RootKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

// Get retrieves a RootKey by ID.
func (s *MemoryStore) Get(id string) (RootKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.keys[id]
	if !ok {
		return RootKey{}, ErrKeyNotFound
	}
	return k, nil
}

// Delete removes a RootKey from the store.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

// Update replaces an existing RootKey.
func (s *MemoryStore) Update(k RootKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; !ok {
		return ErrKeyNotFound
	}
	s.keys[k.ID] = k
	return nil
}

// Create inserts a root key into the database.
func (s *SQLStore) Create(k RootKey) error {
	if err := s.db.Create(&k).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrKeyExists
		}
		return err
	}
	return nil
}

// Get retrieves a root key by ID.
func (s *SQLStore) Get(id string) (RootKey, error) {
	var k RootKey
	if err := s.db.First(&k, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return RootKey{}, ErrKeyNotFound
		}
		return RootKey{}, err
	}
	return k, nil
}

// Delete removes a root key.
func (s *SQLStore) Delete(id string) error {
	res := s.db.Delete(&RootKey{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Update replaces a root key.
func (s *SQLStore) Update(k RootKey) error {
	res := s.db.Model(&RootKey{}).Where("id = ?", k.ID).Updates(k)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Error definitions for store operations.
var (
	ErrKeyNotFound = errors.New("root key not found")
	ErrKeyExists   = errors.New("root key already exists")
)
