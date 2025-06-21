package orgs

// Roles define permission levels within an organization.
const (
	// RoleOwner has full administrative control over the organization and its members.
	RoleOwner = "owner"
	// RoleAdmin can manage organization resources but cannot remove or demote owners.
	RoleAdmin = "admin"
	// RoleMember is a regular member with access limited to permitted resources.
	RoleMember = "member"
)

var allowedRoles = map[string]struct{}{
	RoleOwner:  {},
	RoleAdmin:  {},
	RoleMember: {},
}

// ValidateRole returns true if r matches one of the allowed roles.
func ValidateRole(r string) bool {
	_, ok := allowedRoles[r]
	return ok
}
