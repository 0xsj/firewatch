package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/oauth"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// OAuthHandler handles OAuth authentication flows.
type OAuthHandler struct {
	oauthLoginCmd  *command.OAuthLoginCommand
	stateManager   *oauth.StateManager
	oauthProviders map[string]oauth.Provider
	logger         logger.Logger
}

// NewOAuthHandler creates a new OAuth handler.
func NewOAuthHandler(
	oauthLoginCmd *command.OAuthLoginCommand,
	stateManager *oauth.StateManager,
	oauthProviders map[string]oauth.Provider,
	logger logger.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		oauthLoginCmd:  oauthLoginCmd,
		stateManager:   stateManager,
		oauthProviders: oauthProviders,
		logger:         logger,
	}
}

// InitiateOAuth godoc
// @Summary      Initiate OAuth flow
// @Description  Redirects to the OAuth provider's authorization page. After authorization, the provider will redirect back to the callback endpoint.
// @Tags         identity
// @Param        provider query string true "OAuth provider name" Enums(google, github)
// @Param        tenant_id query string false "Tenant ID for new user registration" format(uuid)
// @Success      307 "Redirect to OAuth provider"
// @Failure      400 {object} ErrorResponse "Provider is required or unsupported"
// @Failure      500 {object} ErrorResponse "Failed to initiate OAuth"
// @Router       /api/v1/auth/oauth [get]
func (h *OAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		response.BadRequest(w, "provider is required")
		return
	}

	// Get provider
	provider, exists := h.oauthProviders[providerName]
	if !exists {
		response.BadRequest(w, "unsupported OAuth provider")
		return
	}

	// Generate state token for CSRF protection
	state, err := h.stateManager.Generate()
	if err != nil {
		h.logger.Error("failed to generate state token", logger.Err(err))
		response.InternalServerError(w, "failed to initiate OAuth")
		return
	}

	// Store tenant_id in session/cookie if provided (for new user registration)
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID != "" {
		// TODO: Store tenant_id in secure cookie or session
		// For now, we'll require it in the callback URL
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_tenant_id",
			Value:    tenantID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   600, // 10 minutes
		})
	}

	// Get authorization URL
	authURL := provider.AuthCodeURL(state)

	h.logger.Info("OAuth flow initiated",
		logger.String("provider", providerName),
	)

	// Redirect to provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuthCallback godoc
// @Summary      OAuth callback
// @Description  Handles the callback from the OAuth provider after user authorization. Exchanges the authorization code for tokens and logs the user in.
// @Tags         identity
// @Produce      json
// @Param        provider query string true "OAuth provider name" Enums(google, github)
// @Param        code query string true "Authorization code from OAuth provider"
// @Param        state query string true "State token for CSRF validation"
// @Success      200 {object} dto.LoginResponse "Login successful with tokens"
// @Failure      400 {object} ErrorResponse "Invalid state, missing code, or unsupported provider"
// @Failure      401 {object} ErrorResponse "OAuth authentication failed"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/v1/auth/oauth/callback [get]
func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		response.BadRequest(w, "provider is required")
		return
	}

	// Validate state token
	state := r.URL.Query().Get("state")
	if err := h.stateManager.Validate(state); err != nil {
		h.logger.Warn("invalid OAuth state", logger.String("error", err.Error()))
		response.BadRequest(w, "invalid or expired OAuth state")
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		h.logger.Warn("missing OAuth code")
		response.BadRequest(w, "missing authorization code")
		return
	}

	// Get tenant_id from cookie (if set during initiation)
	tenantID := ""
	if cookie, err := r.Cookie("oauth_tenant_id"); err == nil {
		tenantID = cookie.Value
		// Clear the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_tenant_id",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
	}

	// Handle OAuth login
	cmdReq := command.OAuthLoginRequest{
		Provider:  providerName,
		Code:      code,
		TenantID:  tenantID,
		IPAddress: extractIPAddress(r),
		UserAgent: r.Header.Get("User-Agent"),
	}

	loginResp, err := h.oauthLoginCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("OAuth login failed",
			logger.String("provider", providerName),
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	h.logger.Info("OAuth login successful",
		logger.String("provider", providerName),
		logger.String("user_id", loginResp.User.ID),
	)

	// Return tokens as JSON
	// In production, you might redirect to a frontend URL with tokens
	RespondWithLogin(w, loginResp)
}

// Ensure dto is used (for swag to find types)
var _ dto.LoginResponse
