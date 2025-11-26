package v1

import (
	"net/http"

	"github.com/0xsj/hexagonal-go/internal/tenant/application/dto"
	"github.com/0xsj/hexagonal-go/pkg/http/response"
)

// RespondWithTenant writes a tenant DTO as JSON response.
func RespondWithTenant(w http.ResponseWriter, statusCode int, tenant *dto.TenantDTO) {
	response.JSON(w, statusCode, tenant)
}

// RespondWithTenantList writes a paginated list of tenants.
func RespondWithTenantList(w http.ResponseWriter, list *dto.TenantListDTO) {
	response.JSON(w, http.StatusOK, list)
}

// RespondWithSettings writes tenant settings as JSON response.
func RespondWithSettings(w http.ResponseWriter, settings *dto.TenantSettingsDTO) {
	response.JSON(w, http.StatusOK, settings)
}

// RespondWithMessage writes a simple success message.
func RespondWithMessage(w http.ResponseWriter, message string) {
	response.JSON(w, http.StatusOK, map[string]string{
		"message": message,
	})
}

// RespondCreated writes a 201 Created response with tenant data.
func RespondCreated(w http.ResponseWriter, tenant *dto.TenantDTO) {
	response.JSON(w, http.StatusCreated, tenant)
}

// RespondNoContent writes a 204 No Content response.
func RespondNoContent(w http.ResponseWriter) {
	response.NoContent(w)
}
