package users

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// Store defines the persistence behavior for User objects.
type Store interface {
	Create(User) error
	Get(id string) (User, error)
	GetByAPIKey(key string) (User, error)
	Delete(id string) error
	Update(User) error
}

// MemoryStore holds users in memory with concurrency safety.
type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]User
	byKey map[string]User
}

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]User), byKey: make(map[string]User)}
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *gorm.DB) *PostgresStore {
	db.AutoMigrate(&User{})
	return &PostgresStore{db: db}
}

// PostgresStore persists users in a PostgreSQL database.
type PostgresStore struct {
	db *gorm.DB
}

// Create inserts a new User. Returns error if ID already exists.
func (s *MemoryStore) Create(u User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.ID]; ok {
		return ErrUserExists
	}
	s.users[u.ID] = u
	s.byKey[u.APIKey] = u
	return nil
}

// Get retrieves a User by ID.
func (s *MemoryStore) Get(id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

// GetByAPIKey retrieves a User by its API key.
func (s *MemoryStore) GetByAPIKey(key string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byKey[key]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

// Delete removes a User.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(s.users, id)
	delete(s.byKey, u.APIKey)
	return nil
}

// Update replaces an existing user.
func (s *MemoryStore) Update(u User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.ID]; !ok {
		return ErrUserNotFound
	}
	old := s.users[u.ID]
	delete(s.byKey, old.APIKey)
	s.users[u.ID] = u
	s.byKey[u.APIKey] = u
	return nil
}

// Create inserts a new user into the database.
func (s *PostgresStore) Create(u User) error {
	if err := s.db.Create(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

// Get retrieves a user by ID.
func (s *PostgresStore) Get(id string) (User, error) {
	var u User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return u, nil
}

// GetByAPIKey retrieves a user by API key.
func (s *PostgresStore) GetByAPIKey(key string) (User, error) {
	var u User
	if err := s.db.First(&u, "api_key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return u, nil
}

// Delete removes a user by ID.
func (s *PostgresStore) Delete(id string) error {
	res := s.db.Delete(&User{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Update modifies an existing user.
func (s *PostgresStore) Update(u User) error {
	res := s.db.Model(&User{}).Where("id = ?", u.ID).Updates(u)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)
