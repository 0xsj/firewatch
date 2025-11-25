package dto

// CreateTemplateResponse is the output for creating a template.
type CreateTemplateResponse struct {
	Template TemplateDTO `json:"template"`
}

// UpdateTemplateResponse is the output for updating a template.
type UpdateTemplateResponse struct {
	Template TemplateDTO `json:"template"`
}

// ActivateTemplateResponse is the output for activating a template.
type ActivateTemplateResponse struct {
	Template TemplateDTO `json:"template"`
}

// ArchiveTemplateResponse is the output for archiving a template.
type ArchiveTemplateResponse struct {
	Template TemplateDTO `json:"template"`
}

// DeleteTemplateResponse is the output for deleting a template.
type DeleteTemplateResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// GetTemplateResponse is the output for getting a template.
type GetTemplateResponse struct {
	Template TemplateDTO `json:"template"`
}

// ListTemplatesResponse is the output for listing templates.
type ListTemplatesResponse struct {
	Templates []TemplateDTO `json:"templates"`
	Total     int           `json:"total"`
	Limit     int           `json:"limit"`
	Offset    int           `json:"offset"`
	HasMore   bool          `json:"has_more"`
}

// PreviewTemplateResponse is the output for previewing a template.
type PreviewTemplateResponse struct {
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text,omitempty"`
}
