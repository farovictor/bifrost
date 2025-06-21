package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/farovictor/bifrost/pkg/auth"
	routes "github.com/farovictor/bifrost/routes"
)

// orgCtxKey is the context key for organization context.
type orgCtxKey struct{}

// OrgContext holds authentication context about the requester.
type OrgContext struct {
	UserID string
	OrgID  string
	Role   string
}

// OrgFromContext extracts organization context from ctx.
func OrgFromContext(ctx context.Context) OrgContext {
	v, _ := ctx.Value(orgCtxKey{}).(OrgContext)
	return v
}

// OrgCtxMiddleware validates the auth token and stores membership info in context.
func OrgCtxMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimPrefix(authHeader, "Bearer ")
			tok, err := auth.Verify(raw)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			role := ""
			if m, err := routes.MembershipStore.Get(tok.UserID, tok.OrgID); err == nil {
				role = m.Role
			}

			ctx := context.WithValue(r.Context(), orgCtxKey{}, OrgContext{
				UserID: tok.UserID,
				OrgID:  tok.OrgID,
				Role:   role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
