package rootkeys

import (
	"database/sql"
	"errors"
	"sync"
)

// Store defines persistence behavior for RootKey objects.
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

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]RootKey)}
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// PostgresStore persists RootKeys in PostgreSQL.
type PostgresStore struct {
	db *sql.DB
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

// Create inserts a root key into the database.
func (s *PostgresStore) Create(k RootKey) error {
	_, err := s.db.Exec("INSERT INTO root_keys (id, api_key) VALUES ($1,$2)", k.ID, k.APIKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrKeyExists
		}
	}
	return err
}

// Get retrieves a root key by ID.
func (s *PostgresStore) Get(id string) (RootKey, error) {
	var k RootKey
	row := s.db.QueryRow("SELECT id, api_key FROM root_keys WHERE id=$1", id)
	if err := row.Scan(&k.ID, &k.APIKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RootKey{}, ErrKeyNotFound
		}
		return RootKey{}, err
	}
	return k, nil
}

// Delete removes a root key.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM root_keys WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Update replaces a root key.
func (s *PostgresStore) Update(k RootKey) error {
	res, err := s.db.Exec("UPDATE root_keys SET api_key=$1 WHERE id=$2", k.APIKey, k.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Error definitions for store operations.
var (
	ErrKeyNotFound = errors.New("root key not found")
	ErrKeyExists   = errors.New("root key already exists")
)
