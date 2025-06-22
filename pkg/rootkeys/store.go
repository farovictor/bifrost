package rootkeys

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
)

// Store defines operations for persisting root keys.
type Store interface {
	Create(RootKey) error
	Get(id string) (RootKey, error)
	Delete(id string) error
	Update(RootKey) error
}

// MemoryStore keeps RootKeys in memory with concurrency safety.
type MemoryStore struct {
	mu   sync.RWMutex
	keys map[string]RootKey
}

// NewStore creates an initialized in-memory Store.
func NewStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]RootKey)}
}

// Create inserts a new RootKey. Returns error if ID already exists.
func (s *MemoryStore) Create(k RootKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

// Get retrieves a RootKey by ID.
func (s *MemoryStore) Get(id string) (RootKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.keys[id]
	if !ok {
		return RootKey{}, ErrKeyNotFound
	}
	return k, nil
}

// Delete removes a RootKey from the store.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

// Update replaces an existing RootKey.
func (s *MemoryStore) Update(k RootKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; !ok {
		return ErrKeyNotFound
	}
	s.keys[k.ID] = k
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
func (s *PostgresStore) Create(k RootKey) error {
	_, err := s.db.Exec(`INSERT INTO root_keys (id, api_key) VALUES ($1,$2)`, k.ID, k.APIKey)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return ErrKeyExists
	}
	return err
}

// Get fetches a root key by id.
func (s *PostgresStore) Get(id string) (RootKey, error) {
	var rk RootKey
	err := s.db.QueryRow(`SELECT id, api_key FROM root_keys WHERE id=$1`, id).Scan(&rk.ID, &rk.APIKey)
	if err == sql.ErrNoRows {
		return RootKey{}, ErrKeyNotFound
	}
	return rk, err
}

// Delete removes a row.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM root_keys WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Update replaces an existing row.
func (s *PostgresStore) Update(k RootKey) error {
	res, err := s.db.Exec(`UPDATE root_keys SET api_key=$2 WHERE id=$1`, k.ID, k.APIKey)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

var (
	ErrKeyNotFound = errors.New("root key not found")
	ErrKeyExists   = errors.New("root key already exists")
)
