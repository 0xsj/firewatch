package dto

import (
	"time"

	"github.com/0xsj/hexagonal-go/internal/email/domain"
)

// TemplateDTO represents an email template for API responses.
type TemplateDTO struct {
	ID          string        `json:"id"`
	TenantID    *string       `json:"tenant_id,omitempty"`
	Slug        string        `json:"slug"`
	Locale      string        `json:"locale"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Subject     string        `json:"subject"`
	BodyHTML    string        `json:"body_html"`
	BodyText    string        `json:"body_text,omitempty"`
	Variables   []VariableDTO `json:"variables"`
	Status      string        `json:"status"`
	Version     int           `json:"version"`
	CreatedBy   *string       `json:"created_by,omitempty"`
	UpdatedBy   *string       `json:"updated_by,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// VariableDTO represents a template variable for API responses.
type VariableDTO struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Required    bool    `json:"required"`
	Default     *string `json:"default,omitempty"`
	Description string  `json:"description,omitempty"`
}

// TemplateListDTO represents a paginated list of templates.
type TemplateListDTO struct {
	Templates []TemplateDTO `json:"templates"`
	Total     int           `json:"total"`
	Limit     int           `json:"limit"`
	Offset    int           `json:"offset"`
	HasMore   bool          `json:"has_more"`
}

// TemplatePreviewDTO represents a rendered template preview.
type TemplatePreviewDTO struct {
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text,omitempty"`
}

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
