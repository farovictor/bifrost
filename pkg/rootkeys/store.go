package rootkeys

import (
	"errors"
	"sync"
)

// Store keeps RootKeys in memory with concurrency safety.
type Store struct {
	mu   sync.RWMutex
	keys map[string]RootKey
}

// NewStore creates an initialized Store.
func NewStore() *Store {
	return &Store{keys: make(map[string]RootKey)}
}

// Create inserts a new RootKey. Returns error if ID already exists.
func (s *Store) Create(k RootKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

// Get retrieves a RootKey by ID.
func (s *Store) Get(id string) (RootKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.keys[id]
	if !ok {
		return RootKey{}, ErrKeyNotFound
	}
	return k, nil
}

// Delete removes a RootKey from the store.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

// Update replaces an existing RootKey.
func (s *Store) Update(k RootKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; !ok {
		return ErrKeyNotFound
	}
	s.keys[k.ID] = k
	return nil
}

// Error definitions for store operations.
var (
	ErrKeyNotFound = errors.New("root key not found")
	ErrKeyExists   = errors.New("root key already exists")
)
