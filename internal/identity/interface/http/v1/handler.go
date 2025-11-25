package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/query"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/http/middleware"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Handler handles HTTP requests for the Identity domain (v1 API).
type Handler struct {
	registerUserCmd         *command.RegisterUserCommand
	loginCmd                *command.LoginCommand
	logoutCmd               *command.LogoutCommand
	refreshTokenCmd         *command.RefreshTokenCommand
	verifyEmailCmd          *command.VerifyEmailCommand
	requestPasswordResetCmd *command.RequestPasswordResetCommand
	resetPasswordCmd        *command.ResetPasswordCommand
	getUserQuery            *query.GetUserQuery
	getCurrentUserQuery     *query.GetCurrentUserQuery
	listUsersQuery          *query.ListUsersQuery
	listSessionsQuery       *query.ListSessionsQuery
	logger                  logger.Logger
}

// NewHandler creates a new v1 identity HTTP handler.
func NewHandler(
	registerUserCmd *command.RegisterUserCommand,
	loginCmd *command.LoginCommand,
	logoutCmd *command.LogoutCommand,
	refreshTokenCmd *command.RefreshTokenCommand,
	verifyEmailCmd *command.VerifyEmailCommand,
	requestPasswordResetCmd *command.RequestPasswordResetCommand,
	resetPasswordCmd *command.ResetPasswordCommand,
	getUserQuery *query.GetUserQuery,
	getCurrentUserQuery *query.GetCurrentUserQuery,
	listUsersQuery *query.ListUsersQuery,
	listSessionsQuery *query.ListSessionsQuery,
	log logger.Logger,
) *Handler {
	return &Handler{
		registerUserCmd:         registerUserCmd,
		loginCmd:                loginCmd,
		logoutCmd:               logoutCmd,
		refreshTokenCmd:         refreshTokenCmd,
		verifyEmailCmd:          verifyEmailCmd,
		requestPasswordResetCmd: requestPasswordResetCmd,
		resetPasswordCmd:        resetPasswordCmd,
		getUserQuery:            getUserQuery,
		getCurrentUserQuery:     getCurrentUserQuery,
		listUsersQuery:          listUsersQuery,
		listSessionsQuery:       listSessionsQuery,
		logger:                  log,
	}
}

// Register handles user registration with password.
// POST /api/v1/users/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseRegisterRequest(r)
	if err != nil {
		h.logger.Warn("invalid registration request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.RegisterRequest{
		TenantID: dtoReq.TenantID,
		Email:    dtoReq.Email,
		Password: dtoReq.Password,
		Role:     user.RoleUser,
	}

	userDTO, err := h.registerUserCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("user registration failed", logger.String("email", cmdReq.Email), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user registered", logger.String("user_id", userDTO.ID), logger.String("email", userDTO.Email))
	RespondCreated(w, userDTO)
}

// Login handles user login with email and password.
// POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseLoginRequest(r)
	if err != nil {
		h.logger.Warn("invalid login request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.LoginRequest{
		Email:     dtoReq.Email,
		Password:  dtoReq.Password,
		IPAddress: dtoReq.IPAddress,
		UserAgent: dtoReq.UserAgent,
	}

	loginResp, err := h.loginCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Warn("login failed", logger.String("email", cmdReq.Email), logger.String("ip_address", cmdReq.IPAddress), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user logged in", logger.String("user_id", loginResp.User.ID), logger.String("email", loginResp.User.Email), logger.String("ip_address", cmdReq.IPAddress))
	RespondWithLogin(w, loginResp)
}

// Logout handles user logout.
// POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract claims from context (set by auth middleware)
	sessionIDStr := middleware.GetSessionID(r.Context())
	if sessionIDStr == "" {
		h.logger.Error("missing session_id in context")
		response.InternalServerError(w, "invalid session")
		return
	}

	sessionID, err := types.ParseID(sessionIDStr)
	if err != nil {
		h.logger.Error("invalid session_id", logger.String("session_id", sessionIDStr), logger.Err(err))
		response.BadRequest(w, "invalid session")
		return
	}

	// Get access token from Authorization header
	accessToken := extractBearerToken(r)

	cmdReq := command.LogoutRequest{
		SessionID:   sessionID,
		AccessToken: accessToken,
	}

	err = h.logoutCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("logout failed", logger.String("session_id", sessionIDStr), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user logged out", logger.String("session_id", sessionIDStr))
	RespondWithMessage(w, "Logged out successfully")
}

// RefreshToken handles token refresh.
// POST /api/v1/auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseRefreshTokenRequest(r)
	if err != nil {
		h.logger.Warn("invalid refresh token request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.RefreshTokenRequest{
		RefreshToken: dtoReq.RefreshToken,
	}

	refreshResp, err := h.refreshTokenCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Warn("token refresh failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("token refreshed")
	response.JSON(w, http.StatusOK, refreshResp)
}

// VerifyEmail handles email verification.
// POST /api/v1/users/verify-email
// GET /api/v1/users/verify-email?token=...
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseVerifyEmailRequest(r)
	if err != nil {
		h.logger.Warn("invalid email verification request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.VerifyEmailRequest{
		Token: dtoReq.Token,
	}

	err = h.verifyEmailCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("email verification failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("email verified")
	RespondWithMessage(w, "Email verified successfully. Your account is now active.")
}

// GetUser retrieves a user by ID.
// GET /api/v1/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := ParseUserID(r)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	userDTO, err := h.getUserQuery.Handle(r.Context(), userID)
	if err != nil {
		h.logger.Warn("get user failed", logger.String("user_id", userID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user retrieved", logger.String("user_id", userDTO.ID))
	RespondWithUser(w, http.StatusOK, userDTO)
}

// GetCurrentUser retrieves the currently authenticated user.
// GET /api/v1/users/me
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Extract user_id from JWT claims (set by auth middleware)
	userIDStr := middleware.GetUserID(r.Context())
	if userIDStr == "" {
		h.logger.Error("missing user_id in context")
		response.InternalServerError(w, "invalid user")
		return
	}

	userID, err := types.ParseID(userIDStr)
	if err != nil {
		h.logger.Error("invalid user_id", logger.String("user_id", userIDStr), logger.Err(err))
		response.BadRequest(w, "invalid user")
		return
	}

	queryReq := query.GetCurrentUserRequest{
		UserID: userID,
	}

	userDTO, err := h.getCurrentUserQuery.Handle(r.Context(), queryReq)
	if err != nil {
		h.logger.Error("get current user failed", logger.String("user_id", userIDStr), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("current user retrieved", logger.String("user_id", userDTO.ID))
	RespondWithUser(w, http.StatusOK, userDTO)
}

// ListUsers retrieves a paginated list of users.
// GET /api/v1/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	req, err := ParseListUsersRequest(r)
	if err != nil {
		h.logger.Warn("invalid list users request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	listResp, err := h.listUsersQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("list users failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("users listed", logger.Int("count", len(listResp.Users)), logger.Int("total", listResp.TotalCount))
	RespondWithUserList(w, listResp)
}

// ListSessions retrieves active sessions for the current user.
// GET /api/v1/sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Extract user_id and tenant_id from JWT claims (set by auth middleware)
	userIDStr := middleware.GetUserID(r.Context())
	tenantID := middleware.GetTenantID(r.Context())

	if userIDStr == "" {
		h.logger.Error("missing user_id in context")
		response.InternalServerError(w, "invalid user")
		return
	}

	userID, err := types.ParseID(userIDStr)
	if err != nil {
		h.logger.Error("invalid user_id", logger.String("user_id", userIDStr), logger.Err(err))
		response.BadRequest(w, "invalid user")
		return
	}

	queryReq := query.ListSessionsRequest{
		UserID:   userID,
		TenantID: tenantID,
	}

	sessions, err := h.listSessionsQuery.Handle(r.Context(), queryReq)
	if err != nil {
		h.logger.Error("list sessions failed", logger.String("user_id", userIDStr), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("sessions listed", logger.String("user_id", userIDStr), logger.Int("count", len(sessions)))
	response.JSON(w, http.StatusOK, map[string]any{
		"sessions": sessions,
	})
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseRequestPasswordResetRequest(r)
	if err != nil {
		h.logger.Warn("invalid password reset request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.RequestPasswordResetRequest{
		Email:     dtoReq.Email,
		IPAddress: dtoReq.IPAddress,
		UserAgent: dtoReq.UserAgent,
	}

	err = h.requestPasswordResetCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("password reset request failed", logger.String("email", cmdReq.Email), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("password reset email sent", logger.String("email", cmdReq.Email))
	RespondWithMessage(w, "If the email exists, a password reset link has been sent")
}

// ResetPassword handles password reset with token.
// POST /api/v1/auth/password/reset
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	dtoReq, err := ParseResetPasswordRequest(r)
	if err != nil {
		h.logger.Warn("invalid reset password request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.ResetPasswordRequest{
		Token:       dtoReq.Token,
		NewPassword: dtoReq.NewPassword,
		IPAddress:   dtoReq.IPAddress,
	}

	err = h.resetPasswordCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("password reset failed", logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("password reset successful")
	RespondWithMessage(w, "Password reset successfully. You can now log in with your new password")
}

// Health handles health check requests.
// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	RespondWithMessage(w, "OK")
}

// extractBearerToken extracts the JWT token from Authorization header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
