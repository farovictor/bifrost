package tests

import (
	"testing"

	"github.com/farovictor/bifrost/pkg/orgs"
	routes "github.com/farovictor/bifrost/routes"
)

func TestCreateGetOrg(t *testing.T) {
	routes.OrgStore = orgs.NewMemoryStore()
	o := orgs.Organization{ID: "org1", Name: "Test Org", Domain: "example.com", Email: "org@example.com"}
	if err := routes.OrgStore.Create(o); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := routes.OrgStore.Get(o.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != o.ID || got.Name != o.Name {
		t.Fatalf("unexpected org: %#v", got)
	}
}

func TestCreateOrgDuplicateName(t *testing.T) {
	routes.OrgStore = orgs.NewMemoryStore()
	o1 := orgs.Organization{ID: "org1", Name: "Dup", Domain: "example.com", Email: "dup1@example.com"}
	if err := routes.OrgStore.Create(o1); err != nil {
		t.Fatalf("seed: %v", err)
	}
	o2 := orgs.Organization{ID: "org2", Name: "Dup", Domain: "example.org", Email: "dup2@example.com"}
	if err := routes.OrgStore.Create(o2); err != orgs.ErrOrgExists {
		t.Fatalf("expected ErrOrgExists, got %v", err)
	}
}

func TestUpdateOrg(t *testing.T) {
	routes.OrgStore = orgs.NewMemoryStore()
	o := orgs.Organization{ID: "org1", Name: "Old", Domain: "example.com", Email: "org@example.com"}
	if err := routes.OrgStore.Create(o); err != nil {
		t.Fatalf("seed: %v", err)
	}
	updated := orgs.Organization{ID: "org1", Name: "New", Domain: "example.com", Email: "org@example.com"}
	if err := routes.OrgStore.Update(updated); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := routes.OrgStore.Get(o.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "New" {
		t.Fatalf("update did not persist")
	}
}

func TestDeleteOrg(t *testing.T) {
	routes.OrgStore = orgs.NewMemoryStore()
	o := orgs.Organization{ID: "org1", Name: "Del", Domain: "example.com", Email: "org@example.com"}
	if err := routes.OrgStore.Create(o); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := routes.OrgStore.Delete(o.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := routes.OrgStore.Get(o.ID); err != orgs.ErrOrgNotFound {
		t.Fatalf("org not removed")
	}
}

func TestCreateGetMembership(t *testing.T) {
	routes.MembershipStore = orgs.NewMembershipStore()
	m := orgs.Membership{UserID: "u1", OrgID: "o1", Role: orgs.RoleMember}
	if err := routes.MembershipStore.Create(m); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := routes.MembershipStore.Get(m.UserID, m.OrgID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != m {
		t.Fatalf("unexpected membership: %#v", got)
	}
}

func TestUpdateMembership(t *testing.T) {
	routes.MembershipStore = orgs.NewMembershipStore()
	m := orgs.Membership{UserID: "u1", OrgID: "o1", Role: orgs.RoleMember}
	if err := routes.MembershipStore.Create(m); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m.Role = orgs.RoleAdmin
	if err := routes.MembershipStore.Update(m); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := routes.MembershipStore.Get(m.UserID, m.OrgID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Role != orgs.RoleAdmin {
		t.Fatalf("update not persisted")
	}
}

func TestDeleteMembership(t *testing.T) {
	routes.MembershipStore = orgs.NewMembershipStore()
	m := orgs.Membership{UserID: "u1", OrgID: "o1", Role: orgs.RoleMember}
	if err := routes.MembershipStore.Create(m); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := routes.MembershipStore.Delete(m.UserID, m.OrgID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := routes.MembershipStore.Get(m.UserID, m.OrgID); err != orgs.ErrMembershipNotFound {
		t.Fatalf("membership not removed")
	}
}
