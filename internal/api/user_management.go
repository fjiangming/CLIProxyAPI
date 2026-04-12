// user_management.go provides user-level OAuth management routes.
// These routes use API Key authentication (same as /v1/* endpoints)
// and expose only OAuth-related endpoints, keeping all other management
// functionality restricted to the management key.
//
// [custom] This file is a custom addition and is not part of the upstream
// CLIProxyAPI repository. When merging upstream updates, this file should
// never conflict.
package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/managementasset"
)

// registerUserManagementRoutes registers user-level OAuth management routes.
// Authentication: API Key (via AuthMiddleware), NOT management key.
// Scope: OAuth operations only — no config, logs, usage, or other admin access.
func (s *Server) registerUserManagementRoutes() {
	if s == nil || s.engine == nil || s.mgmt == nil {
		return
	}

	// Serve the user OAuth page (no auth required for the HTML page itself)
	s.engine.GET("/user-oauth.html", s.serveUserOAuthPage)

	// User-level OAuth API routes (API Key authenticated)
	userOAuth := s.engine.Group("/v0/user-oauth")
	userOAuth.Use(AuthMiddleware(s.accessManager))
	{
		// OAuth: initiate login flows
		userOAuth.GET("/anthropic-auth-url", s.mgmt.RequestAnthropicToken)
		userOAuth.GET("/codex-auth-url", s.mgmt.RequestCodexToken)
		userOAuth.GET("/gemini-cli-auth-url", s.mgmt.RequestGeminiCLIToken)
		userOAuth.GET("/antigravity-auth-url", s.mgmt.RequestAntigravityToken)
		userOAuth.GET("/qwen-auth-url", s.mgmt.RequestQwenToken)
		userOAuth.GET("/kimi-auth-url", s.mgmt.RequestKimiToken)
		userOAuth.GET("/iflow-auth-url", s.mgmt.RequestIFlowToken)
		userOAuth.POST("/iflow-auth-url", s.mgmt.RequestIFlowCookieToken)

		// OAuth: callback and status polling
		userOAuth.POST("/oauth-callback", s.mgmt.PostOAuthCallback)
		userOAuth.GET("/get-auth-status", s.mgmt.GetAuthStatus)

		// Mock config and auth check for management.html bootstrapping
		userOAuth.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"remote-management": gin.H{"allow-remote": true}})
		})
		userOAuth.GET("/check-auth", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Read-only: list authenticated accounts
		userOAuth.GET("/auth-files", s.mgmt.ListAuthFiles)
	}
}

// serveUserOAuthPage serves the user-facing OAuth HTML page from the static directory.
func (s *Server) serveUserOAuthPage(c *gin.Context) {
	staticDir := managementasset.StaticDir(s.configFilePath)
	if strings.TrimSpace(staticDir) == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	filePath := filepath.Join(staticDir, "user-oauth.html")
	if _, err := os.Stat(filePath); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.File(filePath)
}
