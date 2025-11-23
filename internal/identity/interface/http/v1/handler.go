package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/application/command"
	"github.com/0xsj/hexagonal-go/internal/identity/application/query"
	"github.com/0xsj/hexagonal-go/internal/identity/domain/user"
	"github.com/0xsj/hexagonal-go/pkg/observability/logger"
)

// Handler handles HTTP requests for the Identity domain (v1 API).
type Handler struct {
	registerUserCmd *command.RegisterUserCommand
	loginCmd        *command.LoginCommand
	verifyEmailCmd  *command.VerifyEmailCommand
	getUserQuery    *query.GetUserQuery
	listUsersQuery  *query.ListUsersQuery
	logger          logger.Logger
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

// Health handles health check requests.
// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	RespondWithMessage(w, "OK")
}
