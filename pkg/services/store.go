package services

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

// Store defines persistence behavior for Service objects.
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

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{services: make(map[string]Service)}
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *gorm.DB) *PostgresStore {
	db.AutoMigrate(&Service{})
	return &PostgresStore{db: db}
}

// PostgresStore persists services in PostgreSQL.
type PostgresStore struct {
	db *gorm.DB
}

// Create inserts a new Service. Returns an error if the ID already exists.
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

// Create inserts a service into the database.
func (s *PostgresStore) Create(svc Service) error {
	if err := s.db.Create(&svc).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrServiceExists
		}
		return err
	}
	return nil
}

// Get retrieves a service by ID.
func (s *PostgresStore) Get(id string) (Service, error) {
	var svc Service
	if err := s.db.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Service{}, ErrServiceNotFound
		}
		return Service{}, err
	}
	return svc, nil
}

// Delete removes a service.
func (s *PostgresStore) Delete(id string) error {
	res := s.db.Delete(&Service{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrServiceNotFound
	}
	return nil
}

// Error definitions for Store operations.
var (
	ErrServiceNotFound = errors.New("service not found")
	ErrServiceExists   = errors.New("service already exists")
)
