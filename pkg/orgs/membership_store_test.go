package orgs

import (
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestStore(t *testing.T) *SQLMembershipStore {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return NewSQLMembershipStore(db)
}

func TestSQLMembershipStore_AutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if db.Migrator().HasTable(&Membership{}) {
		t.Fatalf("table should not exist before migration")
	}
	NewSQLMembershipStore(db)
	if !db.Migrator().HasTable(&Membership{}) {
		t.Fatalf("table should exist after migration")
	}
}

func TestSQLMembershipStore_Create(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*SQLMembershipStore)
		m       Membership
		wantErr error
	}{
		{
			name:    "success",
			m:       Membership{UserID: "u1", OrgID: "o1", Role: "admin"},
			wantErr: nil,
		},
		{
			name: "duplicate",
			setup: func(s *SQLMembershipStore) {
				_ = s.Create(Membership{UserID: "u1", OrgID: "o1", Role: "admin"})
			},
			m:       Membership{UserID: "u1", OrgID: "o1", Role: "admin"},
			wantErr: ErrMembershipExists,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			err := s.Create(tc.m)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestSQLMembershipStore_Get(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*SQLMembershipStore)
		userID  string
		orgID   string
		wantErr error
	}{
		{
			name: "found",
			setup: func(s *SQLMembershipStore) {
				_ = s.Create(Membership{UserID: "u1", OrgID: "o1", Role: "member"})
			},
			userID:  "u1",
			orgID:   "o1",
			wantErr: nil,
		},
		{
			name:    "not found",
			userID:  "u2",
			orgID:   "o2",
			wantErr: ErrMembershipNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			_, err := s.Get(tc.userID, tc.orgID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestSQLMembershipStore_Delete(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*SQLMembershipStore)
		userID  string
		orgID   string
		wantErr error
	}{
		{
			name: "existing",
			setup: func(s *SQLMembershipStore) {
				_ = s.Create(Membership{UserID: "u1", OrgID: "o1", Role: "member"})
			},
			userID:  "u1",
			orgID:   "o1",
			wantErr: nil,
		},
		{
			name:    "missing",
			userID:  "u2",
			orgID:   "o2",
			wantErr: ErrMembershipNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			err := s.Delete(tc.userID, tc.orgID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestSQLMembershipStore_Update(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(*SQLMembershipStore)
		m       Membership
		wantErr error
	}{
		{
			name: "success",
			setup: func(s *SQLMembershipStore) {
				_ = s.Create(Membership{UserID: "u1", OrgID: "o1", Role: "member"})
			},
			m:       Membership{UserID: "u1", OrgID: "o1", Role: "admin"},
			wantErr: nil,
		},
		{
			name:    "missing",
			m:       Membership{UserID: "u2", OrgID: "o2", Role: "member"},
			wantErr: ErrMembershipNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			err := s.Update(tc.m)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
			if tc.wantErr == nil {
				got, _ := s.Get(tc.m.UserID, tc.m.OrgID)
				if got.Role != tc.m.Role {
					t.Fatalf("expected role %s, got %s", tc.m.Role, got.Role)
				}
			}
		})
	}
}

func TestSQLMembershipStore_List(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(*SQLMembershipStore)
		expect int
	}{
		{
			name:   "empty",
			expect: 0,
		},
		{
			name: "multiple",
			setup: func(s *SQLMembershipStore) {
				_ = s.Create(Membership{UserID: "u1", OrgID: "o1", Role: "member"})
				_ = s.Create(Membership{UserID: "u2", OrgID: "o1", Role: "admin"})
			},
			expect: 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			got := s.List()
			if len(got) != tc.expect {
				t.Fatalf("expected %d memberships, got %d", tc.expect, len(got))
			}
		})
	}
}

func TestSQLMembershipStore_ListByUser(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(*SQLMembershipStore)
		userID string
		expect int
	}{
		{
			name:   "none",
			userID: "u1",
			expect: 0,
		},
		{
			name: "filtered",
			setup: func(s *SQLMembershipStore) {
				_ = s.Create(Membership{UserID: "u1", OrgID: "o1", Role: "member"})
				_ = s.Create(Membership{UserID: "u1", OrgID: "o2", Role: "admin"})
				_ = s.Create(Membership{UserID: "u2", OrgID: "o1", Role: "member"})
			},
			userID: "u1",
			expect: 2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			if tc.setup != nil {
				tc.setup(s)
			}
			got := s.ListByUser(tc.userID)
			if len(got) != tc.expect {
				t.Fatalf("expected %d memberships, got %d", tc.expect, len(got))
			}
		})
	}
}
