package serviceaccounts

import (
	"sync"

	"gorm.io/gorm"
)

// Store defines persistence behaviour for service accounts.
type Store interface {
	Create(ServiceAccount) error
	Get(id string) (ServiceAccount, error)
	GetByAPIKey(apiKey string) (ServiceAccount, error)
	List() []ServiceAccount
	Delete(id string) error
}

// MemoryStore is an in-memory Store used in tests and in-memory mode.
type MemoryStore struct {
	mu       sync.RWMutex
	accounts []ServiceAccount
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (s *MemoryStore) Create(sa ServiceAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.accounts {
		if a.ID == sa.ID {
			return ErrServiceAccountExists
		}
	}
	s.accounts = append(s.accounts, sa)
	return nil
}

func (s *MemoryStore) Get(id string) (ServiceAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return ServiceAccount{}, ErrServiceAccountNotFound
}

func (s *MemoryStore) GetByAPIKey(apiKey string) (ServiceAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.APIKey == apiKey {
			return a, nil
		}
	}
	return ServiceAccount{}, ErrServiceAccountNotFound
}

func (s *MemoryStore) List() []ServiceAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ServiceAccount, len(s.accounts))
	copy(out, s.accounts)
	return out
}

func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.accounts {
		if a.ID == id {
			s.accounts = append(s.accounts[:i], s.accounts[i+1:]...)
			return nil
		}
	}
	return ErrServiceAccountNotFound
}

// SQLStore persists service accounts in a SQL database via GORM.
type SQLStore struct {
	db *gorm.DB
}

func NewSQLStore(db *gorm.DB) *SQLStore {
	db.AutoMigrate(&ServiceAccount{})
	return &SQLStore{db: db}
}

func (s *SQLStore) Create(sa ServiceAccount) error {
	result := s.db.Create(&sa)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return ErrServiceAccountExists
		}
		return result.Error
	}
	return nil
}

func (s *SQLStore) Get(id string) (ServiceAccount, error) {
	var sa ServiceAccount
	if err := s.db.First(&sa, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ServiceAccount{}, ErrServiceAccountNotFound
		}
		return ServiceAccount{}, err
	}
	return sa, nil
}

func (s *SQLStore) GetByAPIKey(apiKey string) (ServiceAccount, error) {
	var sa ServiceAccount
	if err := s.db.First(&sa, "api_key = ?", apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ServiceAccount{}, ErrServiceAccountNotFound
		}
		return ServiceAccount{}, err
	}
	return sa, nil
}

func (s *SQLStore) List() []ServiceAccount {
	var accounts []ServiceAccount
	s.db.Find(&accounts)
	return accounts
}

func (s *SQLStore) Delete(id string) error {
	result := s.db.Delete(&ServiceAccount{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrServiceAccountNotFound
	}
	return nil
}

// isUniqueViolation reports whether err looks like a unique-constraint violation
// across both SQLite and PostgreSQL.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "UNIQUE constraint failed") || // SQLite
		contains(msg, "duplicate key value") // PostgreSQL
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
