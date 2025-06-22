package routes

import "github.com/farovictor/bifrost/pkg/orgs"

// OrgStore holds defined organizations in memory.
var OrgStore = orgs.NewMemoryStore()

// MembershipStore holds organization memberships in memory.
var MembershipStore = orgs.NewMembershipStore()
