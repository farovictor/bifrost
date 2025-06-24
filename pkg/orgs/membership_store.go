package orgs

import (
	"errors"
	"sync"
)

// MembershipStore defines persistence behavior for Membership objects.
type MembershipStore interface {
	Create(Membership) error
	Get(userID, orgID string) (Membership, error)
	Delete(userID, orgID string) error
	Update(Membership) error
	List() []Membership
	ListByUser(userID string) []Membership
}

// MemoryMembershipStore keeps memberships in memory with concurrency safety.
type MemoryMembershipStore struct {
	mu          sync.RWMutex
	memberships map[string]Membership
}

// NewMemoryMembershipStore creates an initialized MemoryMembershipStore.
func NewMemoryMembershipStore() *MemoryMembershipStore {
	return &MemoryMembershipStore{memberships: make(map[string]Membership)}
}

func membershipKey(userID, orgID string) string {
	return userID + ":" + orgID
}

// Create inserts a new Membership. Returns error if the user/org pair already exists.
func (s *MemoryMembershipStore) Create(m Membership) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := membershipKey(m.UserID, m.OrgID)
	if _, ok := s.memberships[k]; ok {
		return ErrMembershipExists
	}
	s.memberships[k] = m
	return nil
}

// Get retrieves a Membership by user and organization IDs.
func (s *MemoryMembershipStore) Get(userID, orgID string) (Membership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k := membershipKey(userID, orgID)
	m, ok := s.memberships[k]
	if !ok {
		return Membership{}, ErrMembershipNotFound
	}
	return m, nil
}

// Delete removes a Membership from the store.
func (s *MemoryMembershipStore) Delete(userID, orgID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := membershipKey(userID, orgID)
	if _, ok := s.memberships[k]; !ok {
		return ErrMembershipNotFound
	}
	delete(s.memberships, k)
	return nil
}

// Update replaces an existing Membership.
func (s *MemoryMembershipStore) Update(m Membership) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := membershipKey(m.UserID, m.OrgID)
	if _, ok := s.memberships[k]; !ok {
		return ErrMembershipNotFound
	}
	s.memberships[k] = m
	return nil
}

// List returns all Memberships currently in the store.
func (s *MemoryMembershipStore) List() []Membership {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Membership, 0, len(s.memberships))
	for _, m := range s.memberships {
		out = append(out, m)
	}
	return out
}

// ListByUser returns memberships belonging to userID.
func (s *MemoryMembershipStore) ListByUser(userID string) []Membership {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Membership, 0)
	for _, m := range s.memberships {
		if m.UserID == userID {
			out = append(out, m)
		}
	}
	return out
}

var (
	ErrMembershipNotFound = errors.New("membership not found")
	ErrMembershipExists   = errors.New("membership already exists")
)
