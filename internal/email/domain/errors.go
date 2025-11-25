package domain

import (
	pkgerrors "github.com/0xsj/hexagonal-go/pkg/errors"
)

// Domain error codes for email templates.
const (
	ErrCodeTemplateNotFound       pkgerrors.Code = "EMAIL_TEMPLATE_NOT_FOUND"
	ErrCodeTemplateAlreadyExists  pkgerrors.Code = "EMAIL_TEMPLATE_ALREADY_EXISTS"
	ErrCodeTemplateInvalidStatus  pkgerrors.Code = "EMAIL_TEMPLATE_INVALID_STATUS"
	ErrCodeTemplateCannotActivate pkgerrors.Code = "EMAIL_TEMPLATE_CANNOT_ACTIVATE"
	ErrCodeTemplateCannotArchive  pkgerrors.Code = "EMAIL_TEMPLATE_CANNOT_ARCHIVE"
	ErrCodeTemplateCannotEdit     pkgerrors.Code = "EMAIL_TEMPLATE_CANNOT_EDIT"
	ErrCodeTemplateCannotDelete   pkgerrors.Code = "EMAIL_TEMPLATE_CANNOT_DELETE"
	ErrCodeTemplateRenderFailed   pkgerrors.Code = "EMAIL_TEMPLATE_RENDER_FAILED"
	ErrCodeTemplateMissingVar     pkgerrors.Code = "EMAIL_TEMPLATE_MISSING_VARIABLE"
	ErrCodeTemplateInvalidVar     pkgerrors.Code = "EMAIL_TEMPLATE_INVALID_VARIABLE"
)

// ErrTemplateNotFound is returned when a template is not found.
func ErrTemplateNotFound(slug string, locale Locale) error {
	return pkgerrors.NotFound("email.template", "template").
		WithCode(ErrCodeTemplateNotFound).
		WithMeta("slug", slug).
		WithMeta("locale", locale.String())
}

// ErrTemplateAlreadyExists is returned when a template with the same slug/locale exists.
func ErrTemplateAlreadyExists(slug string, locale Locale) error {
	return pkgerrors.Conflict("email.template", "template").
		WithCode(ErrCodeTemplateAlreadyExists).
		WithMeta("slug", slug).
		WithMeta("locale", locale.String())
}

// ErrTemplateCannotActivate is returned when a template cannot be activated.
func ErrTemplateCannotActivate(status Status) error {
	return pkgerrors.Validation("email.template", "template cannot be activated from current status").
		WithCode(ErrCodeTemplateCannotActivate).
		WithMeta("current_status", status.String())
}

// ErrTemplateCannotArchive is returned when a template cannot be archived.
func ErrTemplateCannotArchive(status Status) error {
	return pkgerrors.Validation("email.template", "template cannot be archived from current status").
		WithCode(ErrCodeTemplateCannotArchive).
		WithMeta("current_status", status.String())
}

// ErrTemplateCannotEdit is returned when a template cannot be edited.
func ErrTemplateCannotEdit(status Status) error {
	return pkgerrors.Validation("email.template", "template cannot be edited in current status").
		WithCode(ErrCodeTemplateCannotEdit).
		WithMeta("current_status", status.String())
}

// ErrTemplateCannotDelete is returned when a template cannot be deleted.
func ErrTemplateCannotDelete(status Status) error {
	return pkgerrors.Validation("email.template", "active templates cannot be deleted").
		WithCode(ErrCodeTemplateCannotDelete).
		WithMeta("current_status", status.String())
}

// ErrTemplateRenderFailed is returned when template rendering fails.
func ErrTemplateRenderFailed(err error) error {
	return pkgerrors.Internal("email.template", err).
		WithCode(ErrCodeTemplateRenderFailed)
}

// ErrTemplateMissingVariable is returned when a required variable is missing.
func ErrTemplateMissingVariable(varName string) error {
	return pkgerrors.Validation("email.template", "missing required variable").
		WithCode(ErrCodeTemplateMissingVar).
		WithMeta("variable", varName)
}

// ErrTemplateInvalidVariable is returned when a variable value is invalid.
func ErrTemplateInvalidVariable(varName string, reason string) error {
	return pkgerrors.Validation("email.template", "invalid variable value").
		WithCode(ErrCodeTemplateInvalidVar).
		WithMeta("variable", varName).
		WithMeta("reason", reason)
}
