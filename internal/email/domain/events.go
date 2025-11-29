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
// Aligns with messaging.DomainEvent for consistent event publishing.
type TemplateEvent interface {
	Type() string
	EventTime() time.Time
	AggregateID() types.ID
	AggregateTenantID() string
	Payload() map[string]any
	Version() int
}

// baseTemplateEvent contains common fields for all template events.
type baseTemplateEvent struct {
	eventType  string
	templateID types.ID
	tenantID   string
	eventTime  time.Time
	version    int
}

func (e baseTemplateEvent) Type() string              { return e.eventType }
func (e baseTemplateEvent) EventTime() time.Time      { return e.eventTime }
func (e baseTemplateEvent) AggregateID() types.ID     { return e.templateID }
func (e baseTemplateEvent) AggregateTenantID() string { return e.tenantID }
func (e baseTemplateEvent) Version() int              { return e.version }

// ============================================================================
// Template Created Event
// ============================================================================

// TemplateCreatedEvent is emitted when a template is created.
type TemplateCreatedEvent struct {
	baseTemplateEvent
	Slug      string
	Name      string
	Locale    Locale
	CreatedBy *types.ID
}

func (e TemplateCreatedEvent) Payload() map[string]any {
	payload := map[string]any{
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
			eventType:  EventTemplateCreated,
			templateID: templateID,
			tenantID:   tenantID,
			eventTime:  time.Now(),
			version:    1,
		},
		Slug:      slug,
		Name:      name,
		Locale:    locale,
		CreatedBy: createdBy,
	}
}

// ============================================================================
// Template Updated Event
// ============================================================================

// TemplateUpdatedEvent is emitted when a template is updated.
type TemplateUpdatedEvent struct {
	baseTemplateEvent
	Slug      string
	Changes   []string // List of changed fields
	UpdatedBy *types.ID
}

func (e TemplateUpdatedEvent) Payload() map[string]any {
	payload := map[string]any{
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
			eventType:  EventTemplateUpdated,
			templateID: templateID,
			tenantID:   tenantID,
			eventTime:  time.Now(),
			version:    1,
		},
		Slug:      slug,
		Changes:   changes,
		UpdatedBy: updatedBy,
	}
}

// ============================================================================
// Template Activated Event
// ============================================================================

// TemplateActivatedEvent is emitted when a template is activated.
type TemplateActivatedEvent struct {
	baseTemplateEvent
	Slug        string
	Locale      Locale
	ActivatedBy *types.ID
}

func (e TemplateActivatedEvent) Payload() map[string]any {
	payload := map[string]any{
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
			eventType:  EventTemplateActivated,
			templateID: templateID,
			tenantID:   tenantID,
			eventTime:  time.Now(),
			version:    1,
		},
		Slug:        slug,
		Locale:      locale,
		ActivatedBy: activatedBy,
	}
}

// ============================================================================
// Template Archived Event
// ============================================================================

// TemplateArchivedEvent is emitted when a template is archived.
type TemplateArchivedEvent struct {
	baseTemplateEvent
	Slug       string
	Locale     Locale
	ArchivedBy *types.ID
}

func (e TemplateArchivedEvent) Payload() map[string]any {
	payload := map[string]any{
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
			eventType:  EventTemplateArchived,
			templateID: templateID,
			tenantID:   tenantID,
			eventTime:  time.Now(),
			version:    1,
		},
		Slug:       slug,
		Locale:     locale,
		ArchivedBy: archivedBy,
	}
}

// ============================================================================
// Template Deleted Event
// ============================================================================

// TemplateDeletedEvent is emitted when a template is deleted.
type TemplateDeletedEvent struct {
	baseTemplateEvent
	Slug      string
	Locale    Locale
	DeletedBy *types.ID
}

func (e TemplateDeletedEvent) Payload() map[string]any {
	payload := map[string]any{
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
			eventType:  EventTemplateDeleted,
			templateID: templateID,
			tenantID:   tenantID,
			eventTime:  time.Now(),
			version:    1,
		},
		Slug:      slug,
		Locale:    locale,
		DeletedBy: deletedBy,
	}
}
