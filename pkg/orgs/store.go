package orgs

import (
	"errors"
	"sync"
)

// Store keeps Organizations in memory with concurrency safety.
type Store struct {
	mu   sync.RWMutex
	orgs map[string]Organization
}

// NewStore creates an initialized Store.
func NewStore() *Store {
	return &Store{orgs: make(map[string]Organization)}
}

// Create inserts a new Organization. Returns error if ID already exists.
func (s *Store) Create(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[o.ID]; ok {
		return ErrOrgExists
	}
	s.orgs[o.ID] = o
	return nil
}

// Get retrieves an Organization by ID.
func (s *Store) Get(id string) (Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orgs[id]
	if !ok {
		return Organization{}, ErrOrgNotFound
	}
	return o, nil
}

// Delete removes an Organization from the store.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[id]; !ok {
		return ErrOrgNotFound
	}
	delete(s.orgs, id)
	return nil
}

// Update replaces an existing Organization.
func (s *Store) Update(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[o.ID]; !ok {
		return ErrOrgNotFound
	}
	s.orgs[o.ID] = o
	return nil
}

// List returns all Organizations currently in the store.
func (s *Store) List() []Organization {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Organization, 0, len(s.orgs))
	for _, o := range s.orgs {
		out = append(out, o)
	}
	return out
}

var (
	ErrOrgNotFound = errors.New("organization not found")
	ErrOrgExists   = errors.New("organization already exists")
)
