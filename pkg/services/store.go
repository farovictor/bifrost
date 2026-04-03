package services

import (
	"errors"
	"sync"

	"gorm.io/gorm"

	"github.com/farovictor/bifrost/pkg/database"
)

// Store defines persistence behavior for Service objects.
type Store interface {
	Create(Service) error
	Get(id string) (Service, error)
	Update(Service) error
	Delete(id string) error
	List() []Service
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

// NewSQLStore creates a SQL-backed store.
func NewSQLStore(db *gorm.DB) *SQLStore {
	db.AutoMigrate(&Service{})
	return &SQLStore{db: db}
}

// SQLStore persists services in a SQL database.
type SQLStore struct {
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

// Update replaces a Service.
func (s *MemoryStore) Update(svc Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[svc.ID]; !ok {
		return ErrServiceNotFound
	}
	s.services[svc.ID] = svc
	return nil
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

// List returns all Services currently in the store.
func (s *MemoryStore) List() []Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Service, 0, len(s.services))
	for _, svc := range s.services {
		out = append(out, svc)
	}
	return out
}

// Create inserts a service into the database.
func (s *SQLStore) Create(svc Service) error {
	if err := s.db.Create(&svc).Error; err != nil {
		if database.IsDuplicateError(err) {
			return ErrServiceExists
		}
		return err
	}
	return nil
}

// Get retrieves a service by ID.
func (s *SQLStore) Get(id string) (Service, error) {
	var svc Service
	if err := s.db.First(&svc, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Service{}, ErrServiceNotFound
		}
		return Service{}, err
	}
	return svc, nil
}

// Update replaces a service.
func (s *SQLStore) Update(svc Service) error {
	res := s.db.Where("id = ?", svc.ID).Updates(&svc)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrServiceNotFound
	}
	return nil
}

// Delete removes a service.
func (s *SQLStore) Delete(id string) error {
	res := s.db.Delete(&Service{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrServiceNotFound
	}
	return nil
}

// List returns all services from the database.
func (s *SQLStore) List() []Service {
	var out []Service
	if err := s.db.Find(&out).Error; err != nil {
		return nil
	}
	return out
}

// Error definitions for Store operations.
var (
	ErrServiceNotFound = errors.New("service not found")
	ErrServiceExists   = errors.New("service already exists")
)
