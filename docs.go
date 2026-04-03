// @title Bifrost API
// @version 1.0
// @description Secure, delegated API access proxy. Maps short-lived virtual keys to real credentials and transparently proxies requests to upstream services.
// @contact.name Bifrost
// @license.name MIT
// @host localhost:3333
// @BasePath /
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description User API key. Required together with BearerAuth for most management endpoints.
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token. Format: "Bearer <token>"
package main
