package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/query"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Handler handles HTTP requests for the Identity domain (v1 API).
// It orchestrates application commands/queries and translates between HTTP and domain.
//
// Responsibilities:
//   - Parse HTTP requests
//   - Call application layer (commands/queries)
//   - Format HTTP responses
//   - Handle errors
//
// Does NOT contain:
//   - Business logic (that's in domain)
//   - Use case orchestration (that's in application layer)
//   - Database queries (that's in infrastructure)
type Handler struct {
	// Commands (write operations)
	registerUserCmd *command.RegisterUserCommand
	loginCmd        *command.LoginCommand
	verifyEmailCmd  *command.VerifyEmailCommand

	// Queries (read operations)
	getUserQuery   *query.GetUserQuery
	listUsersQuery *query.ListUsersQuery

	logger logger.Logger
}

// NewHandler creates a new v1 identity HTTP handler.
func NewHandler(
	registerUserCmd *command.RegisterUserCommand,
	loginCmd *command.LoginCommand,
	verifyEmailCmd *command.VerifyEmailCommand,
	getUserQuery *query.GetUserQuery,
	listUsersQuery *query.ListUsersQuery,
	log logger.Logger,
) *Handler {
	return &Handler{
		registerUserCmd: registerUserCmd,
		loginCmd:        loginCmd,
		verifyEmailCmd:  verifyEmailCmd,
		getUserQuery:    getUserQuery,
		listUsersQuery:  listUsersQuery,
		logger:          log,
	}
}

// ============================================================================
// User Registration & Authentication
// ============================================================================

// Register handles user registration with password.
// POST /api/v1/users/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request
	req, err := ParseRegisterRequest(r)
	if err != nil {
		h.logger.Warn("invalid registration request",
			logger.String("error", err.Error()),
		)
		RespondValidationError(w, err)
		return
	}

	// 2. TODO: Extract tenant from authenticated context (when we add auth)
	// For now, use tenant from request body
	// req.TenantID = context.TenantID(r.Context())

	// 3. Call command
	user, err := h.registerUserCmd.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("user registration failed",
			logger.String("email", req.Email),
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	// 4. Log success
	h.logger.Info("user registered",
		logger.String("user_id", user.ID),
		logger.String("email", user.Email),
	)

	// 5. Return response
	RespondCreated(w, user)
}

// Login handles user login with email and password.
// POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request (extracts IP address and user agent)
	req, err := ParseLoginRequest(r)
	if err != nil {
		h.logger.Warn("invalid login request",
			logger.String("error", err.Error()),
		)
		RespondValidationError(w, err)
		return
	}

	// 2. Call command
	loginResp, err := h.loginCmd.Handle(r.Context(), req)
	if err != nil {
		h.logger.Warn("login failed",
			logger.String("email", req.Email),
			logger.String("ip_address", req.IPAddress),
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	// 3. Log success
	h.logger.Info("user logged in",
		logger.String("user_id", loginResp.User.ID),
		logger.String("email", loginResp.User.Email),
		logger.String("ip_address", req.IPAddress),
	)

	// 4. Return response with tokens
	RespondWithLogin(w, loginResp)
}

// VerifyEmail handles email verification.
// POST /api/v1/users/verify-email
// GET /api/v1/users/verify-email?token=...
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request (accepts query param or JSON body)
	req, err := ParseVerifyEmailRequest(r)
	if err != nil {
		h.logger.Warn("invalid email verification request",
			logger.String("error", err.Error()),
		)
		RespondValidationError(w, err)
		return
	}

	// 2. Call command
	msgResp, err := h.verifyEmailCmd.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("email verification failed",
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	// 3. Log success
	h.logger.Info("email verified")

	// 4. Return response
	RespondWithMessage(w, msgResp.Message)
}

// ============================================================================
// User Queries (Read Operations)
// ============================================================================

// GetUser retrieves a user by ID.
// GET /api/v1/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	// 1. Parse user ID from URL
	userID, err := ParseUserID(r)
	if err != nil {
		h.logger.Warn("invalid user ID",
			logger.String("error", err.Error()),
		)
		RespondValidationError(w, err)
		return
	}

	// 2. Call query
	user, err := h.getUserQuery.Handle(r.Context(), userID)
	if err != nil {
		h.logger.Warn("get user failed",
			logger.String("user_id", userID.String()),
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	// 3. Log success
	h.logger.Info("user retrieved",
		logger.String("user_id", user.ID),
	)

	// 4. Return response
	RespondWithUser(w, http.StatusOK, user)
}

// ListUsers retrieves a paginated list of users.
// GET /api/v1/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// 1. Parse query parameters
	req, err := ParseListUsersRequest(r)
	if err != nil {
		h.logger.Warn("invalid list users request",
			logger.String("error", err.Error()),
		)
		RespondValidationError(w, err)
		return
	}

	// 2. Call query
	listResp, err := h.listUsersQuery.Handle(r.Context(), req)
	if err != nil {
		h.logger.Error("list users failed",
			logger.Err(err),
		)
		HandleError(w, err)
		return
	}

	// 3. Log success
	h.logger.Info("users listed",
		logger.Int("count", len(listResp.Users)),
		logger.Int("total", listResp.TotalCount),
	)

	// 4. Return response
	RespondWithUserList(w, listResp)
}

// ============================================================================
// Health Check
// ============================================================================

// Health handles health check requests.
// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	RespondWithMessage(w, "OK")
}
