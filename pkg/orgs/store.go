package orgs

import (
	"database/sql"
	"errors"
	"sync"
)

// Store defines persistence behavior for Organization objects.
type Store interface {
	Create(Organization) error
	Get(id string) (Organization, error)
	Delete(id string) error
	Update(Organization) error
	List() []Organization
}

// MemoryStore keeps Organizations in memory with concurrency safety.
type MemoryStore struct {
	mu   sync.RWMutex
	orgs map[string]Organization
}

// NewMemoryStore creates an initialized MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{orgs: make(map[string]Organization)}
}

// NewPostgresStore creates a Postgres-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// PostgresStore persists organizations in PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// Create inserts a new Organization. Returns error if ID already exists.
func (s *MemoryStore) Create(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[o.ID]; ok {
		return ErrOrgExists
	}
	s.orgs[o.ID] = o
	return nil
}

// Get retrieves an Organization by ID.
func (s *MemoryStore) Get(id string) (Organization, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orgs[id]
	if !ok {
		return Organization{}, ErrOrgNotFound
	}
	return o, nil
}

// Delete removes an Organization from the store.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[id]; !ok {
		return ErrOrgNotFound
	}
	delete(s.orgs, id)
	return nil
}

// Update replaces an existing Organization.
func (s *MemoryStore) Update(o Organization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orgs[o.ID]; !ok {
		return ErrOrgNotFound
	}
	s.orgs[o.ID] = o
	return nil
}

// List returns all Organizations currently in the store.
func (s *MemoryStore) List() []Organization {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Organization, 0, len(s.orgs))
	for _, o := range s.orgs {
		out = append(out, o)
	}
	return out
}

// Create inserts an organization into the database.
func (s *PostgresStore) Create(o Organization) error {
	_, err := s.db.Exec("INSERT INTO organizations (id, name) VALUES ($1, $2)", o.ID, o.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOrgExists
		}
	}
	return err
}

// Get retrieves an organization by ID.
func (s *PostgresStore) Get(id string) (Organization, error) {
	var o Organization
	row := s.db.QueryRow("SELECT id, name FROM organizations WHERE id=$1", id)
	if err := row.Scan(&o.ID, &o.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Organization{}, ErrOrgNotFound
		}
		return Organization{}, err
	}
	return o, nil
}

// Delete removes an organization.
func (s *PostgresStore) Delete(id string) error {
	res, err := s.db.Exec("DELETE FROM organizations WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrOrgNotFound
	}
	return nil
}

// Update replaces an organization.
func (s *PostgresStore) Update(o Organization) error {
	res, err := s.db.Exec("UPDATE organizations SET name=$1 WHERE id=$2", o.Name, o.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrOrgNotFound
	}
	return nil
}

// List returns all organizations.
func (s *PostgresStore) List() []Organization {
	rows, err := s.db.Query("SELECT id, name FROM organizations")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Organization
	for rows.Next() {
		var o Organization
		if err := rows.Scan(&o.ID, &o.Name); err != nil {
			continue
		}
		out = append(out, o)
	}
	return out
}

var (
	ErrOrgNotFound = errors.New("organization not found")
	ErrOrgExists   = errors.New("organization already exists")
)
