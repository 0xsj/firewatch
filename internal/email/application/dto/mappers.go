// internal/email/application/dto/mappers.go
package dto

import "github.com/0xsj/hexagonal-go/internal/email/domain"

// MapTemplateToDTO maps a domain template to a DTO.
func MapTemplateToDTO(t *domain.Template) TemplateDTO {
	dto := TemplateDTO{
		ID:          t.ID().String(),
		TenantID:    t.TenantID(),
		Slug:        t.Slug(),
		Locale:      t.Locale().String(),
		Name:        t.Name(),
		Description: t.Description(),
		Subject:     t.Subject(),
		BodyHTML:    t.BodyHTML(),
		BodyText:    t.BodyText(),
		Variables:   MapVariablesToDTO(t.Variables()),
		Status:      t.Status().String(),
		Version:     t.Version(),
		CreatedAt:   t.CreatedAt().Time(),
		UpdatedAt:   t.UpdatedAt().Time(),
	}

	if t.CreatedBy() != nil {
		createdBy := t.CreatedBy().String()
		dto.CreatedBy = &createdBy
	}
	if t.UpdatedBy() != nil {
		updatedBy := t.UpdatedBy().String()
		dto.UpdatedBy = &updatedBy
	}

	return dto
}

// MapVariablesToDTO maps domain variables to DTOs.
func MapVariablesToDTO(vars domain.Variables) []VariableDTO {
	dtos := make([]VariableDTO, len(vars))
	for i, v := range vars {
		dtos[i] = VariableDTO{
			Name:        v.Name,
			Type:        v.Type.String(),
			Required:    v.Required,
			Default:     v.Default,
			Description: v.Description,
		}
	}
	return dtos
}

// MapTemplatesToDTO maps a slice of domain templates to DTOs.
func MapTemplatesToDTO(templates []*domain.Template) []TemplateDTO {
	dtos := make([]TemplateDTO, len(templates))
	for i, t := range templates {
		dtos[i] = MapTemplateToDTO(t)
	}
	return dtos
}
