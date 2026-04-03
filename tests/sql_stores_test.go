package tests

// SQL store integration tests using an in-memory SQLite database.
// These exercise the SQLStore implementations that in-memory store tests skip.

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/farovictor/bifrost/pkg/database"
	"github.com/farovictor/bifrost/pkg/keys"
	"github.com/farovictor/bifrost/pkg/orgs"
	"github.com/farovictor/bifrost/pkg/rootkeys"
	"github.com/farovictor/bifrost/pkg/services"
	"github.com/farovictor/bifrost/pkg/users"
)

func sqliteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := database.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sqlite connect: %v", err)
	}
	return db
}

// ── keys SQL store ────────────────────────────────────────────────────────────

func TestSQLKeyStore(t *testing.T) {
	store := keys.NewSQLStore(sqliteDB(t))

	k := keys.VirtualKey{ID: "k1", Scope: keys.ScopeRead, Target: "svc", ExpiresAt: time.Now().Add(time.Hour), RateLimit: 10}

	if err := store.Create(k); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(k.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != k.ID || got.Scope != k.Scope {
		t.Fatalf("unexpected key: %#v", got)
	}

	k.Scope = keys.ScopeWrite
	if err := store.Update(k.ID, k); err != nil {
		t.Fatalf("update: %v", err)
	}
	got2, _ := store.Get(k.ID)
	if got2.Scope != keys.ScopeWrite {
		t.Fatalf("update not persisted")
	}

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 key, got %d", len(list))
	}

	if err := store.Delete(k.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(k.ID); err != keys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after delete")
	}
}

func TestSQLKeyStoreDuplicate(t *testing.T) {
	store := keys.NewSQLStore(sqliteDB(t))
	k := keys.VirtualKey{ID: "dup", Scope: keys.ScopeRead, Target: "svc", ExpiresAt: time.Now().Add(time.Hour), RateLimit: 1}
	store.Create(k)
	if err := store.Create(k); err != keys.ErrKeyExists {
		t.Fatalf("expected ErrKeyExists, got %v", err)
	}
}

func TestSQLKeyStoreNotFound(t *testing.T) {
	store := keys.NewSQLStore(sqliteDB(t))
	if _, err := store.Get("nope"); err != keys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound")
	}
	if err := store.Update("nope", keys.VirtualKey{}); err != keys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound on update")
	}
	if err := store.Delete("nope"); err != keys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound on delete")
	}
}

// ── rootkeys SQL store ────────────────────────────────────────────────────────

func TestSQLRootKeyStore(t *testing.T) {
	store := rootkeys.NewSQLStore(sqliteDB(t))

	rk := rootkeys.RootKey{ID: "rk1", APIKey: "secret"}
	if err := store.Create(rk); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(rk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.APIKey != rk.APIKey {
		t.Fatalf("unexpected apikey")
	}

	rk.APIKey = "newsecret"
	if err := store.Update(rk); err != nil {
		t.Fatalf("update: %v", err)
	}
	got2, _ := store.Get(rk.ID)
	if got2.APIKey != "newsecret" {
		t.Fatalf("update not persisted")
	}

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}

	if err := store.Delete(rk.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(rk.ID); err != rootkeys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound")
	}
}

func TestSQLRootKeyStoreDuplicate(t *testing.T) {
	store := rootkeys.NewSQLStore(sqliteDB(t))
	rk := rootkeys.RootKey{ID: "dup", APIKey: "k"}
	store.Create(rk)
	if err := store.Create(rk); err != rootkeys.ErrKeyExists {
		t.Fatalf("expected ErrKeyExists, got %v", err)
	}
}

func TestSQLRootKeyStoreNotFound(t *testing.T) {
	store := rootkeys.NewSQLStore(sqliteDB(t))
	if _, err := store.Get("nope"); err != rootkeys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound")
	}
	if err := store.Update(rootkeys.RootKey{ID: "nope"}); err != rootkeys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound on update")
	}
	if err := store.Delete("nope"); err != rootkeys.ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound on delete")
	}
}

// ── services SQL store ────────────────────────────────────────────────────────

func TestSQLServiceStore(t *testing.T) {
	store := services.NewSQLStore(sqliteDB(t))

	svc := services.Service{ID: "svc1", Endpoint: "http://x.com", RootKeyID: "rk"}
	if err := store.Create(svc); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(svc.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Endpoint != svc.Endpoint {
		t.Fatalf("unexpected endpoint")
	}

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}

	if err := store.Delete(svc.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(svc.ID); err != services.ErrServiceNotFound {
		t.Fatalf("expected ErrServiceNotFound")
	}
}

func TestSQLServiceStoreDuplicate(t *testing.T) {
	store := services.NewSQLStore(sqliteDB(t))
	svc := services.Service{ID: "dup", Endpoint: "http://x.com", RootKeyID: "rk"}
	store.Create(svc)
	if err := store.Create(svc); err != services.ErrServiceExists {
		t.Fatalf("expected ErrServiceExists, got %v", err)
	}
}

func TestSQLServiceStoreNotFound(t *testing.T) {
	store := services.NewSQLStore(sqliteDB(t))
	if _, err := store.Get("nope"); err != services.ErrServiceNotFound {
		t.Fatalf("expected ErrServiceNotFound")
	}
	if err := store.Delete("nope"); err != services.ErrServiceNotFound {
		t.Fatalf("expected ErrServiceNotFound on delete")
	}
}

// ── users SQL store ───────────────────────────────────────────────────────────

func TestSQLUserStore(t *testing.T) {
	store := users.NewSQLStore(sqliteDB(t))

	u := users.User{ID: "u1", Name: "Alice", Email: "alice@example.com", APIKey: "key1"}
	if err := store.Create(u); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(u.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Email != u.Email {
		t.Fatalf("unexpected email")
	}

	byKey, err := store.GetByAPIKey(u.APIKey)
	if err != nil {
		t.Fatalf("get by api key: %v", err)
	}
	if byKey.ID != u.ID {
		t.Fatalf("wrong user from GetByAPIKey")
	}

	byEmail, err := store.GetByEmail(u.Email)
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if byEmail.ID != u.ID {
		t.Fatalf("wrong user from GetByEmail")
	}
}

func TestSQLUserStoreDuplicate(t *testing.T) {
	store := users.NewSQLStore(sqliteDB(t))
	u := users.User{ID: "dup", Name: "A", Email: "a@a.com", APIKey: "k"}
	store.Create(u)
	if err := store.Create(u); err != users.ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestSQLUserStoreNotFound(t *testing.T) {
	store := users.NewSQLStore(sqliteDB(t))
	if _, err := store.Get("nope"); err != users.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound")
	}
	if _, err := store.GetByAPIKey("nope"); err != users.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound on GetByAPIKey")
	}
	if _, err := store.GetByEmail("nope@nope.com"); err != users.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound on GetByEmail")
	}
}

// ── orgs SQL store ────────────────────────────────────────────────────────────

func TestSQLOrgStore(t *testing.T) {
	store := orgs.NewSQLStore(sqliteDB(t))

	o := orgs.Organization{ID: "o1", Name: "Acme", Domain: "acme.com", Email: "admin@acme.com"}
	if err := store.Create(o); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(o.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != o.Name {
		t.Fatalf("unexpected name")
	}

	o.Name = "Acme Corp"
	if err := store.Update(o); err != nil {
		t.Fatalf("update: %v", err)
	}
	got2, _ := store.Get(o.ID)
	if got2.Name != "Acme Corp" {
		t.Fatalf("update not persisted")
	}

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}

	if err := store.Delete(o.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(o.ID); err != orgs.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound")
	}
}

func TestSQLOrgStoreDuplicate(t *testing.T) {
	store := orgs.NewSQLStore(sqliteDB(t))
	o := orgs.Organization{ID: "dup", Name: "Dup", Domain: "d.com", Email: "d@d.com"}
	store.Create(o)
	if err := store.Create(o); err != orgs.ErrOrgExists {
		t.Fatalf("expected ErrOrgExists, got %v", err)
	}
}

func TestSQLOrgStoreNotFound(t *testing.T) {
	store := orgs.NewSQLStore(sqliteDB(t))
	if _, err := store.Get("nope"); err != orgs.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound")
	}
	if err := store.Update(orgs.Organization{ID: "nope", Name: "X"}); err != orgs.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound on update")
	}
	if err := store.Delete("nope"); err != orgs.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound on delete")
	}
}

// ── membership SQL store ──────────────────────────────────────────────────────

func TestSQLMembershipStore(t *testing.T) {
	store := orgs.NewSQLMembershipStore(sqliteDB(t))

	m := orgs.Membership{UserID: "u1", OrgID: "o1", Role: orgs.RoleMember}
	if err := store.Create(m); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(m.UserID, m.OrgID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Role != m.Role {
		t.Fatalf("unexpected role")
	}

	m.Role = orgs.RoleAdmin
	if err := store.Update(m); err != nil {
		t.Fatalf("update: %v", err)
	}
	got2, _ := store.Get(m.UserID, m.OrgID)
	if got2.Role != orgs.RoleAdmin {
		t.Fatalf("update not persisted")
	}

	byOrg := store.ListByOrg(m.OrgID)
	if len(byOrg) != 1 {
		t.Fatalf("ListByOrg: expected 1, got %d", len(byOrg))
	}

	byUser := store.ListByUser(m.UserID)
	if len(byUser) != 1 {
		t.Fatalf("ListByUser: expected 1, got %d", len(byUser))
	}

	all := store.List()
	if len(all) != 1 {
		t.Fatalf("List: expected 1, got %d", len(all))
	}

	if err := store.Delete(m.UserID, m.OrgID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.Get(m.UserID, m.OrgID); err != orgs.ErrMembershipNotFound {
		t.Fatalf("expected ErrMembershipNotFound")
	}
}

func TestSQLMembershipStoreDuplicate(t *testing.T) {
	store := orgs.NewSQLMembershipStore(sqliteDB(t))
	m := orgs.Membership{UserID: "u1", OrgID: "o1", Role: orgs.RoleMember}
	store.Create(m)
	if err := store.Create(m); err != orgs.ErrMembershipExists {
		t.Fatalf("expected ErrMembershipExists, got %v", err)
	}
}

func TestSQLMembershipStoreNotFound(t *testing.T) {
	store := orgs.NewSQLMembershipStore(sqliteDB(t))
	if _, err := store.Get("nope", "nope"); err != orgs.ErrMembershipNotFound {
		t.Fatalf("expected ErrMembershipNotFound")
	}
	if err := store.Update(orgs.Membership{UserID: "nope", OrgID: "nope"}); err != orgs.ErrMembershipNotFound {
		t.Fatalf("expected ErrMembershipNotFound on update")
	}
	if err := store.Delete("nope", "nope"); err != orgs.ErrMembershipNotFound {
		t.Fatalf("expected ErrMembershipNotFound on delete")
	}
}
