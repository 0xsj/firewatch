package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/identity/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// RespondWithUser writes a user DTO as JSON response.
// Used for endpoints that return a single user (register, get user, etc.)
func RespondWithUser(w http.ResponseWriter, statusCode int, user *dto.UserDTO) {
	response.JSON(w, statusCode, user)
}

// RespondWithLogin writes a login response with user and tokens.
// Used for successful login/authentication endpoints.
func RespondWithLogin(w http.ResponseWriter, loginResp *dto.LoginResponse) {
	response.JSON(w, http.StatusOK, loginResp)
}

// RespondWithUserList writes a paginated list of users.
// Used for list users endpoint.
func RespondWithUserList(w http.ResponseWriter, listResp *dto.ListUsersResponse) {
	response.JSON(w, http.StatusOK, listResp)
}

// RespondWithMessage writes a simple success message.
// Used for operations that don't return data (email verification, password reset, etc.)
func RespondWithMessage(w http.ResponseWriter, message string) {
	response.JSON(w, http.StatusOK, &dto.MessageResponse{
		Message: message,
	})
}

// RespondCreated writes a 201 Created response with user data.
// Used for user registration endpoint.
func RespondCreated(w http.ResponseWriter, user *dto.UserDTO) {
	response.JSON(w, http.StatusCreated, user)
}

// RespondNoContent writes a 204 No Content response.
// Used for successful operations that don't return data (delete, etc.)
func RespondNoContent(w http.ResponseWriter) {
	response.NoContent(w)
}

// RespondAccepted writes a 202 Accepted response.
// Used for async operations (password reset email sent, etc.)
func RespondAccepted(w http.ResponseWriter, message string) {
	response.JSON(w, http.StatusAccepted, &dto.MessageResponse{
		Message: message,
	})
}
