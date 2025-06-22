package services

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
)

// Store defines operations for persisting services.
type Store interface {
	Create(Service) error
	Get(id string) (Service, error)
	Delete(id string) error
}

// MemoryStore provides concurrency-safe storage for Service definitions.
type MemoryStore struct {
	mu       sync.RWMutex
	services map[string]Service
}

// NewStore creates an initialized in-memory Store.
func NewStore() *MemoryStore {
	return &MemoryStore{services: make(map[string]Service)}
}

// Create inserts a new Service.
func (s *MemoryStore) Create(svc Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[svc.ID]; ok {
		return ErrServiceExists
	}
	s.services[svc.ID] = svc
	return nil
}

// Get retrieves a Service by ID.
func (s *MemoryStore) Get(id string) (Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[id]
	if !ok {
		return Service{}, ErrServiceNotFound
	}
	return svc, nil
}

// Delete removes a Service.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[id]; !ok {
		return ErrServiceNotFound
	}
	delete(s.services, id)
	return nil
}

// PostgresStore implements Store backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore returns a Postgres-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Create inserts a new row.
func (s *PostgresStore) Create(svc Service) error {
	_, err := s.db.Exec(`INSERT INTO services (id, endpoint, root_key_id) VALUES ($1,$2,$3)`, svc.ID, svc.Endpoint, svc.RootKeyID)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return ErrServiceExists
	}
	return err
}

// Get retrieves a service.
func (s *PostgresStore) Get(id string) (Service, error) {
	var svc Service
	err := s.db.QueryRow(`SELECT id, endpoint, root_key_id FROM services WHERE id=$1`, id).Scan(&svc.ID, &svc.Endpoint, &svc.RootKeyID)
	if err == sql.ErrNoRows {
		return Service{}, ErrServiceNotFound
	}
	return svc, err
}

// Delete removes a row.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM services WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrServiceNotFound
	}
	return nil
}

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrServiceExists   = errors.New("service already exists")
)
