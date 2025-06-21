package users

import (
	"errors"
	"sync"
)

// Store holds users in memory with concurrency safety.
type Store struct {
	mu    sync.RWMutex
	users map[string]User
	byKey map[string]User
}

// NewStore creates an initialized Store.
func NewStore() *Store {
	return &Store{users: make(map[string]User), byKey: make(map[string]User)}
}

// Create inserts a new User. Returns error if ID already exists.
func (s *Store) Create(u User) error {
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
func (s *Store) Get(id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

// GetByAPIKey retrieves a User by its API key.
func (s *Store) GetByAPIKey(key string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byKey[key]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

// Delete removes a User.
func (s *Store) Delete(id string) error {
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
func (s *Store) Update(u User) error {
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

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)
