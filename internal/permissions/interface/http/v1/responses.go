package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/permissions/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// RespondWithRole writes a role DTO as JSON response.
func RespondWithRole(w http.ResponseWriter, statusCode int, role dto.RoleDTO) {
	response.JSON(w, statusCode, role)
}

// RespondWithRoleList writes a list of roles.
func RespondWithRoleList(w http.ResponseWriter, roles []dto.RoleSummaryDTO, total int) {
	response.JSON(w, http.StatusOK, map[string]any{
		"roles": roles,
		"total": total,
	})
}

// RespondWithAssignment writes an assignment DTO as JSON response.
func RespondWithAssignment(w http.ResponseWriter, statusCode int, assignment dto.AssignmentDTO) {
	response.JSON(w, statusCode, assignment)
}

// RespondWithPermissions writes user permissions as JSON response.
func RespondWithPermissions(w http.ResponseWriter, permissions dto.UserPermissionsDTO) {
	response.JSON(w, http.StatusOK, permissions)
}

// RespondWithPermissionCheck writes a permission check result.
func RespondWithPermissionCheck(w http.ResponseWriter, result dto.PermissionCheckDTO) {
	response.JSON(w, http.StatusOK, result)
}

// RespondWithMessage writes a simple success message.
func RespondWithMessage(w http.ResponseWriter, message string) {
	response.JSON(w, http.StatusOK, map[string]string{
		"message": message,
	})
}

// RespondCreated writes a 201 Created response.
func RespondCreated(w http.ResponseWriter, data any) {
	response.JSON(w, http.StatusCreated, data)
}

// RespondNoContent writes a 204 No Content response.
func RespondNoContent(w http.ResponseWriter) {
	response.NoContent(w)
}
