package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
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
	changePasswordCmd       *command.ChangePasswordCommand
	suspendUserCmd          *command.SuspendUserCommand
	reactivateUserCmd       *command.ReactivateUserCommand
	changeUserRoleCmd       *command.ChangeUserRoleCommand
	deleteUserCmd           *command.DeleteUserCommand
	getUserQuery            *query.GetUserQuery
	getCurrentUserQuery     *query.GetCurrentUserQuery
	listUsersQuery          *query.ListUsersQuery
	listSessionsQuery       *query.ListSessionsQuery
	oauthHandler            *OAuthHandler
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
	changePasswordCmd *command.ChangePasswordCommand,
	suspendUserCmd *command.SuspendUserCommand,
	reactivateUserCmd *command.ReactivateUserCommand,
	changeUserRoleCmd *command.ChangeUserRoleCommand,
	deleteUserCmd *command.DeleteUserCommand,
	getUserQuery *query.GetUserQuery,
	getCurrentUserQuery *query.GetCurrentUserQuery,
	listUsersQuery *query.ListUsersQuery,
	listSessionsQuery *query.ListSessionsQuery,
	oauthHandler *OAuthHandler,
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
		changePasswordCmd:       changePasswordCmd,
		suspendUserCmd:          suspendUserCmd,
		reactivateUserCmd:       reactivateUserCmd,
		changeUserRoleCmd:       changeUserRoleCmd,
		deleteUserCmd:           deleteUserCmd,
		getUserQuery:            getUserQuery,
		getCurrentUserQuery:     getCurrentUserQuery,
		listUsersQuery:          listUsersQuery,
		listSessionsQuery:       listSessionsQuery,
		oauthHandler:            oauthHandler,
		logger:                  log,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with email and password
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterUserRequest true "Registration request"
// @Success      201 {object} dto.UserDTO "User created successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or validation error"
// @Failure      409 {object} ErrorResponse "Email already taken"
// @Router       /api/v1/users/register [post]
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

// Login godoc
// @Summary      User login
// @Description  Authenticates a user with email and password, returns JWT tokens
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200 {object} dto.LoginResponse "Login successful with tokens"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Invalid credentials"
// @Failure      403 {object} ErrorResponse "Account suspended or email not verified"
// @Failure      429 {object} ErrorResponse "Too many login attempts"
// @Router       /api/v1/auth/login [post]
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

// Logout godoc
// @Summary      User logout
// @Description  Invalidates the current session and access token
// @Tags         identity
// @Produce      json
// @Success      200 {object} dto.MessageResponse "Logged out successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/v1/auth/logout [post]
// @Security     BearerAuth
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

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access token
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} dto.LoginResponse "New tokens issued"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Invalid or expired refresh token"
// @Router       /api/v1/auth/refresh [post]
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

// VerifyEmail godoc
// @Summary      Verify email address
// @Description  Verifies a user's email address using the token sent via email
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        token query string false "Verification token (for GET requests)"
// @Param        request body dto.VerifyEmailRequest false "Verification token (for POST requests)"
// @Success      200 {object} dto.MessageResponse "Email verified successfully"
// @Failure      400 {object} ErrorResponse "Invalid or missing token"
// @Failure      404 {object} ErrorResponse "Token not found or expired"
// @Router       /api/v1/users/verify-email [get]
// @Router       /api/v1/users/verify-email [post]
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

// GetUser godoc
// @Summary      Get user by ID
// @Description  Retrieves a user by their unique identifier
// @Tags         identity
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} dto.UserDTO "User found"
// @Failure      400 {object} ErrorResponse "Invalid user ID format"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/v1/users/{id} [get]
// @Security     BearerAuth
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

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Retrieves the currently authenticated user's profile
// @Tags         identity
// @Produce      json
// @Success      200 {object} dto.UserDTO "Current user profile"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/v1/users/me [get]
// @Security     BearerAuth
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

// ListUsers godoc
// @Summary      List users
// @Description  Retrieves a paginated list of users with optional filters
// @Tags         identity
// @Produce      json
// @Param        status query string false "Filter by status" Enums(pending, active, suspended, deleted)
// @Param        role query string false "Filter by role" Enums(guest, user, moderator, admin, super_admin)
// @Param        email_verified query bool false "Filter by email verification status"
// @Param        email_contains query string false "Filter by email containing text"
// @Param        offset query int false "Pagination offset" default(0) minimum(0)
// @Param        limit query int false "Pagination limit" default(50) minimum(1) maximum(100)
// @Param        sort_by query string false "Sort field" Enums(created_at, updated_at, email, last_login_at)
// @Param        sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success      200 {object} dto.ListUsersResponse "List of users"
// @Failure      400 {object} ErrorResponse "Invalid query parameters"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/users [get]
// @Security     BearerAuth
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

// ListSessions godoc
// @Summary      List user sessions
// @Description  Retrieves active sessions for the currently authenticated user
// @Tags         identity
// @Produce      json
// @Success      200 {object} SessionsResponse "List of active sessions"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/v1/sessions [get]
// @Security     BearerAuth
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

// RequestPasswordReset godoc
// @Summary      Request password reset
// @Description  Sends a password reset email to the specified address if it exists
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body dto.RequestPasswordResetRequest true "Password reset request"
// @Success      200 {object} dto.MessageResponse "Reset email sent (if account exists)"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Router       /api/v1/auth/password/forgot [post]
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

// ResetPassword godoc
// @Summary      Reset password
// @Description  Resets the user's password using a valid reset token
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "New password with reset token"
// @Success      200 {object} dto.MessageResponse "Password reset successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or weak password"
// @Failure      404 {object} ErrorResponse "Token not found or expired"
// @Router       /api/v1/auth/password/reset [post]
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

// Health godoc
// @Summary      Health check
// @Description  Returns OK if the service is healthy
// @Tags         system
// @Produce      json
// @Success      200 {object} dto.MessageResponse "Service is healthy"
// @Router       /health [get]
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

// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the authenticated user's password
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        request body dto.ChangePasswordRequest true "Old and new password"
// @Success      200 {object} dto.MessageResponse "Password changed successfully"
// @Failure      400 {object} ErrorResponse "Invalid request body or weak password"
// @Failure      401 {object} ErrorResponse "Unauthorized or incorrect old password"
// @Router       /api/v1/users/me/password [post]
// @Security     BearerAuth
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
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

	dtoReq, err := ParseChangePasswordRequest(r)
	if err != nil {
		h.logger.Warn("invalid change password request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.ChangePasswordRequest{
		UserID:      userID,
		OldPassword: dtoReq.OldPassword,
		NewPassword: dtoReq.NewPassword,
		IPAddress:   dtoReq.IPAddress,
	}

	err = h.changePasswordCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("password change failed", logger.String("user_id", userIDStr), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("password changed successfully", logger.String("user_id", userIDStr))
	RespondWithMessage(w, "Password changed successfully")
}

// SuspendUser godoc
// @Summary      Suspend user
// @Description  Suspends a user account, preventing login. Requires admin privileges.
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        request body dto.SuspendUserRequest true "Suspension reason"
// @Success      200 {object} dto.MessageResponse "User suspended successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/v1/users/{id}/suspend [post]
// @Security     BearerAuth
func (h *Handler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from JWT claims
	adminIDStr := middleware.GetUserID(r.Context())
	if adminIDStr == "" {
		h.logger.Error("missing admin user_id in context")
		response.InternalServerError(w, "invalid admin")
		return
	}

	adminID, err := types.ParseID(adminIDStr)
	if err != nil {
		h.logger.Error("invalid admin user_id", logger.String("admin_id", adminIDStr), logger.Err(err))
		response.BadRequest(w, "invalid admin")
		return
	}

	// Extract target user ID from URL
	userID, err := ParseUserID(r)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	// Parse request body
	dtoReq, err := ParseSuspendUserRequest(r)
	if err != nil {
		h.logger.Warn("invalid suspend user request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.SuspendUserRequest{
		UserID:      userID,
		Reason:      dtoReq.Reason,
		SuspendedBy: adminID,
	}

	err = h.suspendUserCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("user suspension failed", logger.String("user_id", userID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user suspended", logger.String("user_id", userID.String()), logger.String("admin_id", adminIDStr))
	RespondWithMessage(w, "User suspended successfully")
}

// ReactivateUser godoc
// @Summary      Reactivate user
// @Description  Reactivates a suspended user account. Requires admin privileges.
// @Tags         identity
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} dto.MessageResponse "User reactivated successfully"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      422 {object} ErrorResponse "Invalid status transition"
// @Router       /api/v1/users/{id}/reactivate [post]
// @Security     BearerAuth
func (h *Handler) ReactivateUser(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from JWT claims
	adminIDStr := middleware.GetUserID(r.Context())
	if adminIDStr == "" {
		h.logger.Error("missing admin user_id in context")
		response.InternalServerError(w, "invalid admin")
		return
	}

	adminID, err := types.ParseID(adminIDStr)
	if err != nil {
		h.logger.Error("invalid admin user_id", logger.String("admin_id", adminIDStr), logger.Err(err))
		response.BadRequest(w, "invalid admin")
		return
	}

	// Extract target user ID from URL
	userID, err := ParseUserID(r)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.ReactivateUserRequest{
		UserID:        userID,
		ReactivatedBy: adminID,
	}

	err = h.reactivateUserCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("user reactivation failed", logger.String("user_id", userID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user reactivated", logger.String("user_id", userID.String()), logger.String("admin_id", adminIDStr))
	RespondWithMessage(w, "User reactivated successfully")
}

// ChangeUserRole godoc
// @Summary      Change user role
// @Description  Changes a user's role. Requires admin privileges.
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        request body dto.ChangeRoleRequest true "New role"
// @Success      200 {object} dto.MessageResponse "User role changed successfully"
// @Failure      400 {object} ErrorResponse "Invalid request or role"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/v1/users/{id}/role [post]
// @Security     BearerAuth
func (h *Handler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from JWT claims
	adminIDStr := middleware.GetUserID(r.Context())
	if adminIDStr == "" {
		h.logger.Error("missing admin user_id in context")
		response.InternalServerError(w, "invalid admin")
		return
	}

	adminID, err := types.ParseID(adminIDStr)
	if err != nil {
		h.logger.Error("invalid admin user_id", logger.String("admin_id", adminIDStr), logger.Err(err))
		response.BadRequest(w, "invalid admin")
		return
	}

	// Extract target user ID from URL
	userID, err := ParseUserID(r)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	// Parse request body
	dtoReq, err := ParseChangeRoleRequest(r)
	if err != nil {
		h.logger.Warn("invalid change role request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.ChangeUserRoleRequest{
		UserID:    userID,
		NewRole:   user.Role(dtoReq.Role),
		Reason:    dtoReq.Reason,
		ChangedBy: adminID,
	}

	err = h.changeUserRoleCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("role change failed", logger.String("user_id", userID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user role changed", logger.String("user_id", userID.String()), logger.String("admin_id", adminIDStr), logger.String("new_role", dtoReq.Role))
	RespondWithMessage(w, "User role changed successfully")
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Soft-deletes a user account. Requires admin privileges.
// @Tags         identity
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        request body dto.DeleteUserRequest true "Deletion reason"
// @Success      204 "User deleted successfully"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      403 {object} ErrorResponse "Forbidden - admin access required"
// @Failure      404 {object} ErrorResponse "User not found"
// @Router       /api/v1/users/{id} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from JWT claims
	adminIDStr := middleware.GetUserID(r.Context())
	if adminIDStr == "" {
		h.logger.Error("missing admin user_id in context")
		response.InternalServerError(w, "invalid admin")
		return
	}

	adminID, err := types.ParseID(adminIDStr)
	if err != nil {
		h.logger.Error("invalid admin user_id", logger.String("admin_id", adminIDStr), logger.Err(err))
		response.BadRequest(w, "invalid admin")
		return
	}

	// Extract target user ID from URL
	userID, err := ParseUserID(r)
	if err != nil {
		h.logger.Warn("invalid user ID", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	// Parse request body
	dtoReq, err := ParseDeleteUserRequest(r)
	if err != nil {
		h.logger.Warn("invalid delete user request", logger.String("error", err.Error()))
		RespondValidationError(w, err)
		return
	}

	cmdReq := command.DeleteUserRequest{
		UserID:    userID,
		Reason:    dtoReq.Reason,
		DeletedBy: adminID,
	}

	err = h.deleteUserCmd.Handle(r.Context(), cmdReq)
	if err != nil {
		h.logger.Error("user deletion failed", logger.String("user_id", userID.String()), logger.Err(err))
		HandleError(w, err)
		return
	}

	h.logger.Info("user deleted", logger.String("user_id", userID.String()), logger.String("admin_id", adminIDStr))
	RespondNoContent(w)
}

// ============================================================================
// Swagger Models
// ============================================================================

// ErrorResponse represents an error response body.
// @Description Error response returned by the API
type ErrorResponse struct {
	Code    string         `json:"code" example:"USER_NOT_FOUND"`
	Message string         `json:"message" example:"user not found"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// SessionsResponse represents the list sessions response.
// @Description List of user sessions
type SessionsResponse struct {
	Sessions []dto.SessionDTO `json:"sessions"`
}
