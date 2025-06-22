package keys

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
)

// Store defines operations for virtual key persistence.
type Store interface {
	Create(VirtualKey) error
	Get(id string) (VirtualKey, error)
	Update(id string, k VirtualKey) error
	Delete(id string) error
	List() []VirtualKey
}

// MemoryStore is an in-memory repository for VirtualKey objects.
type MemoryStore struct {
	mu   sync.RWMutex
	keys map[string]VirtualKey
}

// NewStore creates an initialized in-memory store.
func NewStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]VirtualKey)}
}

// Create inserts a new VirtualKey.
func (s *MemoryStore) Create(k VirtualKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[k.ID]; ok {
		return ErrKeyExists
	}
	s.keys[k.ID] = k
	return nil
}

// Get retrieves a VirtualKey by its ID.
func (s *MemoryStore) Get(id string) (VirtualKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.keys[id]
	if !ok {
		return VirtualKey{}, ErrKeyNotFound
	}
	return v, nil
}

// Update replaces the VirtualKey stored under the given ID.
func (s *MemoryStore) Update(id string, k VirtualKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	s.keys[id] = k
	return nil
}

// Delete removes a VirtualKey from the store.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.keys[id]; !ok {
		return ErrKeyNotFound
	}
	delete(s.keys, id)
	return nil
}

// List returns all VirtualKeys currently in the store.
func (s *MemoryStore) List() []VirtualKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VirtualKey, 0, len(s.keys))
	for _, v := range s.keys {
		out = append(out, v)
	}
	return out
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
func (s *PostgresStore) Create(k VirtualKey) error {
	_, err := s.db.Exec(`INSERT INTO virtual_keys (id, scope, expires_at, target, rate_limit) VALUES ($1,$2,$3,$4,$5)`, k.ID, k.Scope, k.ExpiresAt, k.Target, k.RateLimit)
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return ErrKeyExists
	}
	return err
}

// Get retrieves a key.
func (s *PostgresStore) Get(id string) (VirtualKey, error) {
	var v VirtualKey
	err := s.db.QueryRow(`SELECT id, scope, expires_at, target, rate_limit FROM virtual_keys WHERE id=$1`, id).Scan(&v.ID, &v.Scope, &v.ExpiresAt, &v.Target, &v.RateLimit)
	if err == sql.ErrNoRows {
		return VirtualKey{}, ErrKeyNotFound
	}
	return v, err
}

// Update modifies a row.
func (s *PostgresStore) Update(id string, k VirtualKey) error {
	res, err := s.db.Exec(`UPDATE virtual_keys SET scope=$2, expires_at=$3, target=$4, rate_limit=$5 WHERE id=$1`, id, k.Scope, k.ExpiresAt, k.Target, k.RateLimit)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Delete removes a key row.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM virtual_keys WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// List returns all keys.
func (s *PostgresStore) List() []VirtualKey {
	rows, err := s.db.Query(`SELECT id, scope, expires_at, target, rate_limit FROM virtual_keys`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []VirtualKey
	for rows.Next() {
		var v VirtualKey
		if err := rows.Scan(&v.ID, &v.Scope, &v.ExpiresAt, &v.Target, &v.RateLimit); err == nil {
			out = append(out, v)
		}
	}
	return out
}

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)
