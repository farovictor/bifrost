# HTTP Middleware

This directory provides reusable middleware used by the router:

- `AuthMiddleware` – checks the `X-API-Key` or `Authorization` header
- `LoggingMiddleware` – logs method, path, status and duration
- `MetricsMiddleware` – records Prometheus metrics when enabled
- `OrgCtxMiddleware` – verifies bearer tokens and stores organization context
- `RateLimitMiddleware` – limits requests per virtual key using Redis with a
  local fallback
