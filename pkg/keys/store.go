package keys

import (
	"database/sql"
	"errors"
	"sync"
)

// Store defines persistence behavior for VirtualKey objects.
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

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]VirtualKey)}
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// PostgresStore persists VirtualKeys in PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// Create inserts a new VirtualKey. Returns an error if the key ID already exists.
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

// Create inserts a virtual key into the database.
func (s *PostgresStore) Create(k VirtualKey) error {
	_, err := s.db.Exec("INSERT INTO virtual_keys (id, scope, expires_at, target, rate_limit) VALUES ($1,$2,$3,$4,$5)",
		k.ID, k.Scope, k.ExpiresAt, k.Target, k.RateLimit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrKeyExists
		}
	}
	return err
}

// Get retrieves a virtual key by ID.
func (s *PostgresStore) Get(id string) (VirtualKey, error) {
	var v VirtualKey
	row := s.db.QueryRow("SELECT id, scope, expires_at, target, rate_limit FROM virtual_keys WHERE id=$1", id)
	if err := row.Scan(&v.ID, &v.Scope, &v.ExpiresAt, &v.Target, &v.RateLimit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return VirtualKey{}, ErrKeyNotFound
		}
		return VirtualKey{}, err
	}
	return v, nil
}

// Update replaces an existing virtual key.
func (s *PostgresStore) Update(id string, k VirtualKey) error {
	res, err := s.db.Exec("UPDATE virtual_keys SET scope=$1, expires_at=$2, target=$3, rate_limit=$4 WHERE id=$5",
		k.Scope, k.ExpiresAt, k.Target, k.RateLimit, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// Delete removes a virtual key by ID.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM virtual_keys WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrKeyNotFound
	}
	return nil
}

// List returns all virtual keys from the database.
func (s *PostgresStore) List() []VirtualKey {
	rows, err := s.db.Query("SELECT id, scope, expires_at, target, rate_limit FROM virtual_keys")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []VirtualKey
	for rows.Next() {
		var v VirtualKey
		if err := rows.Scan(&v.ID, &v.Scope, &v.ExpiresAt, &v.Target, &v.RateLimit); err != nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

// Error values returned by Store operations.
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)
