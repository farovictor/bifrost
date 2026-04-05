package usage

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

// Store defines persistence behaviour for usage events.
type Store interface {
	Record(Event) error
	List(keyID string, from, to time.Time, page, perPage int) ([]Event, int64, error)
	TotalTokens(keyID string) int
}

// MemoryStore keeps events in memory — used in tests and in-memory mode.
type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
	nextID uint
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Record(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	e.ID = s.nextID
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	s.events = append(s.events, e)
	return nil
}

func (s *MemoryStore) List(keyID string, from, to time.Time, page, perPage int) ([]Event, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Event
	for _, e := range s.events {
		if e.KeyID != keyID {
			continue
		}
		if !from.IsZero() && e.Timestamp.Before(from) {
			continue
		}
		if !to.IsZero() && e.Timestamp.After(to) {
			continue
		}
		filtered = append(filtered, e)
	}

	total := int64(len(filtered))
	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * perPage
	if start >= len(filtered) {
		return []Event{}, total, nil
	}
	end := start + perPage
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func (s *MemoryStore) TotalTokens(keyID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, e := range s.events {
		if e.KeyID == keyID {
			total += e.TotalTokens
		}
	}
	return total
}

// SQLStore persists usage events in a SQL database.
type SQLStore struct {
	db *gorm.DB
}

func NewSQLStore(db *gorm.DB) *SQLStore {
	db.AutoMigrate(&Event{})
	return &SQLStore{db: db}
}

func (s *SQLStore) Record(e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	return s.db.Create(&e).Error
}

func (s *SQLStore) List(keyID string, from, to time.Time, page, perPage int) ([]Event, int64, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}

	q := s.db.Model(&Event{}).Where("key_id = ?", keyID)
	if !from.IsZero() {
		q = q.Where("timestamp >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("timestamp <= ?", to)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []Event
	if err := q.Order("timestamp desc").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (s *SQLStore) TotalTokens(keyID string) int {
	var total int
	s.db.Model(&Event{}).
		Where("key_id = ?", keyID).
		Select("COALESCE(SUM(total_tokens), 0)").
		Scan(&total)
	return total
}
