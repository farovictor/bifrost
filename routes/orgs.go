package routes

import "github.com/farovictor/bifrost/pkg/orgs"

// OrgStore provides access to organizations.
var OrgStore orgs.Store

// MembershipStore provides access to organization memberships.
var MembershipStore orgs.MembershipStore = orgs.NewMemoryMembershipStore()
