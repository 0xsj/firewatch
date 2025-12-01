// internal/permissions/application/dto/mappers.go
package dto

import "github.com/0xsj/hexagonal-go/internal/permissions/domain"

// RoleToDTO converts a domain Role to RoleDTO.
func RoleToDTO(role *domain.Role) RoleDTO {
	return RoleDTO{
		ID:          role.ID().String(),
		TenantID:    role.TenantID(),
		Name:        role.Name(),
		Description: role.Description(),
		Permissions: role.Permissions().Strings(),
		IsSystem:    role.IsSystem(),
		CreatedAt:   role.CreatedAt().Time(),
		UpdatedAt:   role.UpdatedAt().Time(),
	}
}

// RoleToSummaryDTO converts a domain Role to RoleSummaryDTO.
func RoleToSummaryDTO(role *domain.Role) RoleSummaryDTO {
	return RoleSummaryDTO{
		ID:          role.ID().String(),
		Name:        role.Name(),
		Description: role.Description(),
		IsSystem:    role.IsSystem(),
		Permissions: len(role.Permissions()),
	}
}

// RolesToDTOs converts a slice of domain Roles to RoleDTOs.
func RolesToDTOs(roles []*domain.Role) []RoleDTO {
	dtos := make([]RoleDTO, len(roles))
	for i, role := range roles {
		dtos[i] = RoleToDTO(role)
	}
	return dtos
}

// RolesToSummaryDTOs converts a slice of domain Roles to RoleSummaryDTOs.
func RolesToSummaryDTOs(roles []*domain.Role) []RoleSummaryDTO {
	dtos := make([]RoleSummaryDTO, len(roles))
	for i, role := range roles {
		dtos[i] = RoleToSummaryDTO(role)
	}
	return dtos
}

// AssignmentToDTO converts a domain Assignment to AssignmentDTO.
func AssignmentToDTO(assignment *domain.Assignment, roleName string) AssignmentDTO {
	dto := AssignmentDTO{
		ID:        assignment.ID().String(),
		UserID:    assignment.UserID().String(),
		RoleID:    assignment.RoleID().String(),
		RoleName:  roleName,
		TenantID:  assignment.TenantID(),
		CreatedAt: assignment.CreatedAt().Time(),
		CreatedBy: assignment.CreatedBy(),
	}

	if assignment.ResourceID() != "" {
		dto.ResourceID = assignment.ResourceID()
	}

	if assignment.ExpiresAt() != nil {
		dto.ExpiresAt = assignment.ExpiresAt()
	}

	return dto
}

// AssignmentsToDTOs converts a slice of domain Assignments to AssignmentDTOs.
func AssignmentsToDTOs(assignments []*domain.Assignment, roleNames map[string]string) []AssignmentDTO {
	dtos := make([]AssignmentDTO, len(assignments))
	for i, assignment := range assignments {
		roleName := roleNames[assignment.RoleID().String()]
		dtos[i] = AssignmentToDTO(assignment, roleName)
	}
	return dtos
}
