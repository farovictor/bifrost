package keys

// Allowed scopes for virtual keys.
const (
	ScopeRead  = "read"
	ScopeWrite = "write"
)

var allowedScopes = map[string]struct{}{
	ScopeRead:  {},
	ScopeWrite: {},
}

// ValidateScope returns true if s matches one of the allowed scopes.
func ValidateScope(s string) bool {
	_, ok := allowedScopes[s]
	return ok
}
