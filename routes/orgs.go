package routes

import "github.com/farovictor/bifrost/pkg/orgs"

// OrgStore provides access to organizations.
var OrgStore orgs.Store

// MembershipStore holds organization memberships in memory.
var MembershipStore = orgs.NewMembershipStore()
