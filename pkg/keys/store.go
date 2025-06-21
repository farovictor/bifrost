package keys

import (
	"errors"
	"sync"
)

// Store is an in-memory repository for VirtualKey objects.
type Store struct {
	mu   sync.RWMutex
	keys map[string]VirtualKey
}

// NewStore creates an initialized Store.
func NewStore() *Store {
	return &Store{keys: make(map[string]VirtualKey)}
}

// Create inserts a new VirtualKey. Returns an error if the key ID already exists.
func (s *Store) Create(k VirtualKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

// Get retrieves a VirtualKey by its ID.
func (s *Store) Get(id string) (VirtualKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.keys[id]
	if !ok {
		return VirtualKey{}, ErrKeyNotFound
	}
	return v, nil
}

// Update replaces the VirtualKey stored under the given ID.
func (s *Store) Update(id string, k VirtualKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	s.keys[id] = k
	return nil
}

// Delete removes a VirtualKey from the store.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

// List returns all VirtualKeys currently in the store.
func (s *Store) List() []VirtualKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VirtualKey, 0, len(s.keys))
	for _, v := range s.keys {
		out = append(out, v)
	}
	return out
}

// Error values returned by Store operations.
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)
