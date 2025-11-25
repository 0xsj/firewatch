package domain

import (
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Event types for email templates.
const (
	EventTemplateCreated   = "email.template.created"
	EventTemplateUpdated   = "email.template.updated"
	EventTemplateActivated = "email.template.activated"
	EventTemplateArchived  = "email.template.archived"
	EventTemplateDeleted   = "email.template.deleted"
)

// TemplateEvent is the base interface for all template domain events.
type TemplateEvent interface {
	EventType() string
	AggregateID() types.ID
	AggregateTenantID() string
	OccurredAt() time.Time
	Payload() map[string]interface{}
}

// baseTemplateEvent contains common fields for all template events.
type baseTemplateEvent struct {
	templateID types.ID
	tenantID   string
	occurredAt time.Time
}

func (e baseTemplateEvent) AggregateID() types.ID     { return e.templateID }
func (e baseTemplateEvent) AggregateTenantID() string { return e.tenantID }
func (e baseTemplateEvent) OccurredAt() time.Time     { return e.occurredAt }

// TemplateCreatedEvent is emitted when a template is created.
type TemplateCreatedEvent struct {
	baseTemplateEvent
	Slug      string
	Name      string
	Locale    Locale
	CreatedBy *types.ID
}

func (e TemplateCreatedEvent) EventType() string { return EventTemplateCreated }

func (e TemplateCreatedEvent) Payload() map[string]interface{} {
	payload := map[string]interface{}{
		"template_id": e.templateID.String(),
		"tenant_id":   e.tenantID,
		"slug":        e.Slug,
		"name":        e.Name,
		"locale":      e.Locale.String(),
	}
	if e.CreatedBy != nil {
		payload["created_by"] = e.CreatedBy.String()
	}
	return payload
}

// NewTemplateCreatedEvent creates a new TemplateCreatedEvent.
func NewTemplateCreatedEvent(templateID types.ID, tenantID string, slug, name string, locale Locale, createdBy *types.ID) TemplateCreatedEvent {
	return TemplateCreatedEvent{
		baseTemplateEvent: baseTemplateEvent{
			templateID: templateID,
			tenantID:   tenantID,
			occurredAt: time.Now(),
		},
		Slug:      slug,
		Name:      name,
		Locale:    locale,
		CreatedBy: createdBy,
	}
}

// TemplateUpdatedEvent is emitted when a template is updated.
type TemplateUpdatedEvent struct {
	baseTemplateEvent
	Slug      string
	Changes   []string // List of changed fields
	UpdatedBy *types.ID
}

func (e TemplateUpdatedEvent) EventType() string { return EventTemplateUpdated }

func (e TemplateUpdatedEvent) Payload() map[string]interface{} {
	payload := map[string]interface{}{
		"template_id": e.templateID.String(),
		"tenant_id":   e.tenantID,
		"slug":        e.Slug,
		"changes":     e.Changes,
	}
	if e.UpdatedBy != nil {
		payload["updated_by"] = e.UpdatedBy.String()
	}
	return payload
}

// NewTemplateUpdatedEvent creates a new TemplateUpdatedEvent.
func NewTemplateUpdatedEvent(templateID types.ID, tenantID string, slug string, changes []string, updatedBy *types.ID) TemplateUpdatedEvent {
	return TemplateUpdatedEvent{
		baseTemplateEvent: baseTemplateEvent{
			templateID: templateID,
			tenantID:   tenantID,
			occurredAt: time.Now(),
		},
		Slug:      slug,
		Changes:   changes,
		UpdatedBy: updatedBy,
	}
}

// TemplateActivatedEvent is emitted when a template is activated.
type TemplateActivatedEvent struct {
	baseTemplateEvent
	Slug        string
	Locale      Locale
	ActivatedBy *types.ID
}

func (e TemplateActivatedEvent) EventType() string { return EventTemplateActivated }

func (e TemplateActivatedEvent) Payload() map[string]interface{} {
	payload := map[string]interface{}{
		"template_id": e.templateID.String(),
		"tenant_id":   e.tenantID,
		"slug":        e.Slug,
		"locale":      e.Locale.String(),
	}
	if e.ActivatedBy != nil {
		payload["activated_by"] = e.ActivatedBy.String()
	}
	return payload
}

// NewTemplateActivatedEvent creates a new TemplateActivatedEvent.
func NewTemplateActivatedEvent(templateID types.ID, tenantID string, slug string, locale Locale, activatedBy *types.ID) TemplateActivatedEvent {
	return TemplateActivatedEvent{
		baseTemplateEvent: baseTemplateEvent{
			templateID: templateID,
			tenantID:   tenantID,
			occurredAt: time.Now(),
		},
		Slug:        slug,
		Locale:      locale,
		ActivatedBy: activatedBy,
	}
}

// TemplateArchivedEvent is emitted when a template is archived.
type TemplateArchivedEvent struct {
	baseTemplateEvent
	Slug       string
	Locale     Locale
	ArchivedBy *types.ID
}

func (e TemplateArchivedEvent) EventType() string { return EventTemplateArchived }

func (e TemplateArchivedEvent) Payload() map[string]interface{} {
	payload := map[string]interface{}{
		"template_id": e.templateID.String(),
		"tenant_id":   e.tenantID,
		"slug":        e.Slug,
		"locale":      e.Locale.String(),
	}
	if e.ArchivedBy != nil {
		payload["archived_by"] = e.ArchivedBy.String()
	}
	return payload
}

// NewTemplateArchivedEvent creates a new TemplateArchivedEvent.
func NewTemplateArchivedEvent(templateID types.ID, tenantID string, slug string, locale Locale, archivedBy *types.ID) TemplateArchivedEvent {
	return TemplateArchivedEvent{
		baseTemplateEvent: baseTemplateEvent{
			templateID: templateID,
			tenantID:   tenantID,
			occurredAt: time.Now(),
		},
		Slug:       slug,
		Locale:     locale,
		ArchivedBy: archivedBy,
	}
}

// TemplateDeletedEvent is emitted when a template is deleted.
type TemplateDeletedEvent struct {
	baseTemplateEvent
	Slug      string
	Locale    Locale
	DeletedBy *types.ID
}

func (e TemplateDeletedEvent) EventType() string { return EventTemplateDeleted }

func (e TemplateDeletedEvent) Payload() map[string]interface{} {
	payload := map[string]interface{}{
		"template_id": e.templateID.String(),
		"tenant_id":   e.tenantID,
		"slug":        e.Slug,
		"locale":      e.Locale.String(),
	}
	if e.DeletedBy != nil {
		payload["deleted_by"] = e.DeletedBy.String()
	}
	return payload
}

// NewTemplateDeletedEvent creates a new TemplateDeletedEvent.
func NewTemplateDeletedEvent(templateID types.ID, tenantID string, slug string, locale Locale, deletedBy *types.ID) TemplateDeletedEvent {
	return TemplateDeletedEvent{
		baseTemplateEvent: baseTemplateEvent{
			templateID: templateID,
			tenantID:   tenantID,
			occurredAt: time.Now(),
		},
		Slug:      slug,
		Locale:    locale,
		DeletedBy: deletedBy,
	}
}
