package services

import (
	"errors"
	"sync"
)

// Store provides concurrency-safe storage for Service definitions.
type Store struct {
	mu       sync.RWMutex
	services map[string]Service
}

// NewStore creates an initialized Store.
func NewStore() *Store {
	return &Store{services: make(map[string]Service)}
}

// Create inserts a new Service. Returns an error if the ID already exists.
func (s *Store) Create(svc Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[svc.ID]; ok {
		return ErrServiceExists
	}
	s.services[svc.ID] = svc
	return nil
}

// Get retrieves a Service by ID.
func (s *Store) Get(id string) (Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[id]
	if !ok {
		return Service{}, ErrServiceNotFound
	}
	return svc, nil
}

// Delete removes a Service.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[id]; !ok {
		return ErrServiceNotFound
	}
	delete(s.services, id)
	return nil
}

// Error definitions for Store operations.
var (
	ErrServiceNotFound = errors.New("service not found")
	ErrServiceExists   = errors.New("service already exists")
)
