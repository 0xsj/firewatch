package email

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	texttemplate "text/template"
)

// RenderedContent represents rendered email content.
type RenderedContent struct {
	Subject  string
	BodyHTML string
	BodyText string
}

// TemplateData is a map of variable names to values for rendering.
type TemplateData map[string]interface{}

// Renderer renders email templates with provided data.
type Renderer struct {
	leftDelim  string
	rightDelim string
}

// RendererOption configures the renderer.
type RendererOption func(*Renderer)

// WithDelimiters sets custom template delimiters.
func WithDelimiters(left, right string) RendererOption {
	return func(r *Renderer) {
		r.leftDelim = left
		r.rightDelim = right
	}
}

// NewRenderer creates a new template renderer.
func NewRenderer(opts ...RendererOption) *Renderer {
	r := &Renderer{
		leftDelim:  "{{",
		rightDelim: "}}",
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Render renders subject, HTML body, and optional text body with the provided data.
func (r *Renderer) Render(subject, bodyHTML, bodyText string, data TemplateData) (*RenderedContent, error) {
	renderedSubject, err := r.RenderText(subject, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	renderedHTML, err := r.RenderHTML(bodyHTML, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render html body: %w", err)
	}

	var renderedText string
	if bodyText != "" {
		renderedText, err = r.RenderText(bodyText, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render text body: %w", err)
		}
	}

	return &RenderedContent{
		Subject:  renderedSubject,
		BodyHTML: renderedHTML,
		BodyText: renderedText,
	}, nil
}

// RenderHTML renders an HTML template string with the provided data.
func (r *Renderer) RenderHTML(tmplStr string, data TemplateData) (string, error) {
	tmpl, err := template.New("email").
		Delims(r.leftDelim, r.rightDelim).
		Funcs(r.htmlFuncMap()).
		Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse html template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute html template: %w", err)
	}

	return buf.String(), nil
}

// RenderText renders a plain text template string with the provided data.
func (r *Renderer) RenderText(tmplStr string, data TemplateData) (string, error) {
	tmpl, err := texttemplate.New("email").
		Delims(r.leftDelim, r.rightDelim).
		Funcs(r.textFuncMap()).
		Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse text template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute text template: %w", err)
	}

	return buf.String(), nil
}

// htmlFuncMap returns template functions for HTML templates.
func (r *Renderer) htmlFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    toTitle,
		"trim":     strings.TrimSpace,
		"default":  defaultValue,
		"safeHTML": safeHTML,
	}
}

// textFuncMap returns template functions for text templates.
func (r *Renderer) textFuncMap() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"title":   toTitle,
		"trim":    strings.TrimSpace,
		"default": defaultValue,
	}
}

// toTitle converts string to title case.
func toTitle(s string) string {
	return strings.Title(strings.ToLower(s))
}

// defaultValue returns the value if non-empty, otherwise the default.
func defaultValue(def, val interface{}) interface{} {
	if val == nil || val == "" {
		return def
	}
	return val
}

// safeHTML marks a string as safe HTML (not escaped).
func safeHTML(s string) template.HTML {
	return template.HTML(s)
}
