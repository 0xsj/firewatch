// internal/email/application/dto/template.go
package dto

import "time"

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
