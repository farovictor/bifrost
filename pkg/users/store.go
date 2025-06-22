package users

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
)

// Store defines the operations for persisting users.
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

// NewStore creates an initialized in-memory store.
func NewStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]User), byKey: make(map[string]User)}
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

// PostgresStore implements Store backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore returns a PostgresStore using the given db handle.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Create inserts a new user row.
func (s *PostgresStore) Create(u User) error {
	_, err := s.db.Exec(`INSERT INTO users (id, api_key) VALUES ($1,$2)`, u.ID, u.APIKey)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return ErrUserExists
		}
	}
	return err
}

// Get retrieves a user by id.
func (s *PostgresStore) Get(id string) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id, api_key FROM users WHERE id=$1`, id).Scan(&u.ID, &u.APIKey)
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	return u, err
}

// GetByAPIKey fetches a user by api_key.
func (s *PostgresStore) GetByAPIKey(key string) (User, error) {
	var u User
	err := s.db.QueryRow(`SELECT id, api_key FROM users WHERE api_key=$1`, key).Scan(&u.ID, &u.APIKey)
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	return u, err
}

// Delete removes a user row.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Update replaces an existing user row.
func (s *PostgresStore) Update(u User) error {
	res, err := s.db.Exec(`UPDATE users SET api_key=$2 WHERE id=$1`, u.ID, u.APIKey)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)
