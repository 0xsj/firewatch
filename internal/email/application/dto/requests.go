package dto

// CreateTemplateRequest is the input for creating a new template.
type CreateTemplateRequest struct {
	TenantID    *string           `json:"tenant_id,omitempty"`
	Slug        string            `json:"slug"`
	Locale      string            `json:"locale"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Subject     string            `json:"subject"`
	BodyHTML    string            `json:"body_html"`
	BodyText    string            `json:"body_text,omitempty"`
	Variables   []VariableRequest `json:"variables,omitempty"`
}

// VariableRequest represents a variable in create/update requests.
type VariableRequest struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Required    bool    `json:"required"`
	Default     *string `json:"default,omitempty"`
	Description string  `json:"description,omitempty"`
}

// UpdateTemplateRequest is the input for updating a template.
type UpdateTemplateRequest struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Subject     *string           `json:"subject,omitempty"`
	BodyHTML    *string           `json:"body_html,omitempty"`
	BodyText    *string           `json:"body_text,omitempty"`
	Variables   []VariableRequest `json:"variables,omitempty"`
}

// ActivateTemplateRequest is the input for activating a template.
type ActivateTemplateRequest struct {
	ID string `json:"id"`
}

// ArchiveTemplateRequest is the input for archiving a template.
type ArchiveTemplateRequest struct {
	ID string `json:"id"`
}

// DeleteTemplateRequest is the input for deleting a template.
type DeleteTemplateRequest struct {
	ID string `json:"id"`
}

// ListTemplatesRequest is the input for listing templates.
type ListTemplatesRequest struct {
	TenantID               *string `json:"tenant_id,omitempty"`
	IncludeSystemTemplates bool    `json:"include_system_templates"`
	Status                 *string `json:"status,omitempty"`
	Locale                 *string `json:"locale,omitempty"`
	SlugContains           string  `json:"slug_contains,omitempty"`
	NameContains           string  `json:"name_contains,omitempty"`
	Limit                  int     `json:"limit,omitempty"`
	Offset                 int     `json:"offset,omitempty"`
	SortBy                 string  `json:"sort_by,omitempty"`
	SortOrder              string  `json:"sort_order,omitempty"`
}

// GetTemplateRequest is the input for getting a template by ID.
type GetTemplateRequest struct {
	ID string `json:"id"`
}

// GetTemplateBySlugRequest is the input for getting a template by slug.
type GetTemplateBySlugRequest struct {
	TenantID *string `json:"tenant_id,omitempty"`
	Slug     string  `json:"slug"`
	Locale   string  `json:"locale"`
}

// PreviewTemplateRequest is the input for previewing a rendered template.
type PreviewTemplateRequest struct {
	ID   string                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

// PreviewTemplateBySlugRequest is the input for previewing by slug.
type PreviewTemplateBySlugRequest struct {
	TenantID *string                `json:"tenant_id,omitempty"`
	Slug     string                 `json:"slug"`
	Locale   string                 `json:"locale"`
	Data     map[string]interface{} `json:"data"`
}
