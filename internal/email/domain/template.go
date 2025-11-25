package domain

import (
	"fmt"
	"time"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Template is the aggregate root for email templates.
type Template struct {
	// Identity
	id       types.ID
	tenantID *string // nil for system-wide templates
	slug     string
	locale   Locale

	// Content
	name        string
	description string
	subject     string
	bodyHTML    string
	bodyText    string
	variables   Variables

	// State
	status  Status
	version int

	// Audit
	createdBy *types.ID
	updatedBy *types.ID
	createdAt types.Timestamp
	updatedAt types.Timestamp

	// Domain events
	events []TemplateEvent
}

// ============================================================================
// Getters
// ============================================================================

func (t *Template) ID() types.ID               { return t.id }
func (t *Template) TenantID() *string          { return t.tenantID }
func (t *Template) Slug() string               { return t.slug }
func (t *Template) Locale() Locale             { return t.locale }
func (t *Template) Name() string               { return t.name }
func (t *Template) Description() string        { return t.description }
func (t *Template) Subject() string            { return t.subject }
func (t *Template) BodyHTML() string           { return t.bodyHTML }
func (t *Template) BodyText() string           { return t.bodyText }
func (t *Template) Variables() Variables       { return t.variables }
func (t *Template) Status() Status             { return t.status }
func (t *Template) Version() int               { return t.version }
func (t *Template) CreatedBy() *types.ID       { return t.createdBy }
func (t *Template) UpdatedBy() *types.ID       { return t.updatedBy }
func (t *Template) CreatedAt() types.Timestamp { return t.createdAt }
func (t *Template) UpdatedAt() types.Timestamp { return t.updatedAt }

// TenantIDString returns the tenant ID as a string, or empty if system-wide.
func (t *Template) TenantIDString() string {
	if t.tenantID == nil {
		return ""
	}
	return *t.tenantID
}

// IsSystemTemplate returns true if this is a system-wide template.
func (t *Template) IsSystemTemplate() bool {
	return t.tenantID == nil
}

// ============================================================================
// Factory Methods
// ============================================================================

// NewTemplate creates a new email template in draft status.
func NewTemplate(
	id types.ID,
	tenantID *string,
	slug string,
	locale Locale,
	name string,
	description string,
	subject string,
	bodyHTML string,
	bodyText string,
	variables Variables,
	createdBy *types.ID,
) (*Template, error) {
	const op = "Template.New"

	// Validate required fields
	if id.IsEmpty() {
		return nil, fmt.Errorf("%s: id is required", op)
	}
	if slug == "" {
		return nil, fmt.Errorf("%s: slug is required", op)
	}
	if err := locale.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if name == "" {
		return nil, fmt.Errorf("%s: name is required", op)
	}
	if subject == "" {
		return nil, fmt.Errorf("%s: subject is required", op)
	}
	if bodyHTML == "" {
		return nil, fmt.Errorf("%s: body_html is required", op)
	}
	if err := variables.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	now := types.Now()

	t := &Template{
		id:          id,
		tenantID:    tenantID,
		slug:        slug,
		locale:      locale,
		name:        name,
		description: description,
		subject:     subject,
		bodyHTML:    bodyHTML,
		bodyText:    bodyText,
		variables:   variables,
		status:      StatusDraft,
		version:     1,
		createdBy:   createdBy,
		updatedBy:   createdBy,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]TemplateEvent, 0),
	}

	// Emit created event
	t.addEvent(NewTemplateCreatedEvent(
		t.id,
		t.TenantIDString(),
		t.slug,
		t.name,
		t.locale,
		createdBy,
	))

	return t, nil
}

// Reconstitute recreates a template from stored state (no validation, no events).
func Reconstitute(
	id types.ID,
	tenantID *string,
	slug string,
	locale Locale,
	name string,
	description string,
	subject string,
	bodyHTML string,
	bodyText string,
	variables Variables,
	status Status,
	version int,
	createdBy *types.ID,
	updatedBy *types.ID,
	createdAt types.Timestamp,
	updatedAt types.Timestamp,
) *Template {
	return &Template{
		id:          id,
		tenantID:    tenantID,
		slug:        slug,
		locale:      locale,
		name:        name,
		description: description,
		subject:     subject,
		bodyHTML:    bodyHTML,
		bodyText:    bodyText,
		variables:   variables,
		status:      status,
		version:     version,
		createdBy:   createdBy,
		updatedBy:   updatedBy,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]TemplateEvent, 0),
	}
}

// ============================================================================
// Behavior Methods
// ============================================================================

// Update updates the template content. Only allowed in draft status.
func (t *Template) Update(
	name string,
	description string,
	subject string,
	bodyHTML string,
	bodyText string,
	variables Variables,
	updatedBy *types.ID,
) error {
	const op = "Template.Update"

	if !t.status.CanEdit() {
		return ErrTemplateCannotEdit(t.status)
	}

	// Track changes for event
	var changes []string

	if name != "" && name != t.name {
		t.name = name
		changes = append(changes, "name")
	}
	if description != t.description {
		t.description = description
		changes = append(changes, "description")
	}
	if subject != "" && subject != t.subject {
		t.subject = subject
		changes = append(changes, "subject")
	}
	if bodyHTML != "" && bodyHTML != t.bodyHTML {
		t.bodyHTML = bodyHTML
		changes = append(changes, "body_html")
	}
	if bodyText != t.bodyText {
		t.bodyText = bodyText
		changes = append(changes, "body_text")
	}
	if variables != nil {
		if err := variables.Validate(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		t.variables = variables
		changes = append(changes, "variables")
	}

	if len(changes) > 0 {
		t.updatedBy = updatedBy
		t.updatedAt = types.Now()
		t.version++

		t.addEvent(NewTemplateUpdatedEvent(
			t.id,
			t.TenantIDString(),
			t.slug,
			changes,
			updatedBy,
		))
	}

	return nil
}

// Activate activates the template, making it available for use.
func (t *Template) Activate(activatedBy *types.ID) error {
	if !t.status.CanActivate() {
		return ErrTemplateCannotActivate(t.status)
	}

	t.status = StatusActive
	t.updatedBy = activatedBy
	t.updatedAt = types.Now()
	t.version++

	t.addEvent(NewTemplateActivatedEvent(
		t.id,
		t.TenantIDString(),
		t.slug,
		t.locale,
		activatedBy,
	))

	return nil
}

// Archive archives the template, making it unavailable for use.
func (t *Template) Archive(archivedBy *types.ID) error {
	if !t.status.CanArchive() {
		return ErrTemplateCannotArchive(t.status)
	}

	t.status = StatusArchived
	t.updatedBy = archivedBy
	t.updatedAt = types.Now()
	t.version++

	t.addEvent(NewTemplateArchivedEvent(
		t.id,
		t.TenantIDString(),
		t.slug,
		t.locale,
		archivedBy,
	))

	return nil
}

// MarkDeleted prepares the template for deletion.
// Only draft and archived templates can be deleted.
func (t *Template) MarkDeleted(deletedBy *types.ID) error {
	if t.status.IsActive() {
		return ErrTemplateCannotDelete(t.status)
	}

	t.addEvent(NewTemplateDeletedEvent(
		t.id,
		t.TenantIDString(),
		t.slug,
		t.locale,
		deletedBy,
	))

	return nil
}

// ============================================================================
// Event Methods
// ============================================================================

// Events returns all uncommitted domain events.
func (t *Template) Events() []TemplateEvent {
	return t.events
}

// ClearEvents clears all uncommitted domain events.
func (t *Template) ClearEvents() {
	t.events = make([]TemplateEvent, 0)
}

// addEvent adds a domain event to the aggregate.
func (t *Template) addEvent(event TemplateEvent) {
	t.events = append(t.events, event)
}

// ============================================================================
// Snapshot (for persistence)
// ============================================================================

// Snapshot represents the template state for persistence.
type Snapshot struct {
	ID          string
	TenantID    string
	Slug        string
	Locale      string
	Name        string
	Description string
	Subject     string
	BodyHTML    string
	BodyText    string
	Variables   []VariableSnapshot
	Status      string
	Version     int
	IsSystem    bool
	CreatedBy   string
	UpdatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ActivatedAt *time.Time
	ArchivedAt  *time.Time
}

// VariableSnapshot represents a variable for persistence.
type VariableSnapshot struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Required    bool    `json:"required"`
	Default     *string `json:"default,omitempty"`
	Description string  `json:"description,omitempty"`
}

// ToSnapshot converts the template to a snapshot for persistence.
func (t *Template) ToSnapshot() Snapshot {
	vars := make([]VariableSnapshot, len(t.variables))
	for i, v := range t.variables {
		vars[i] = VariableSnapshot{
			Name:        v.Name,
			Type:        v.Type.String(),
			Required:    v.Required,
			Default:     v.Default,
			Description: v.Description,
		}
	}

	var tenantID string
	if t.tenantID != nil {
		tenantID = *t.tenantID
	}

	var createdBy, updatedBy string
	if t.createdBy != nil {
		createdBy = t.createdBy.String()
	}
	if t.updatedBy != nil {
		updatedBy = t.updatedBy.String()
	}

	return Snapshot{
		ID:          t.id.String(),
		TenantID:    tenantID,
		Slug:        t.slug,
		Locale:      t.locale.String(),
		Name:        t.name,
		Description: t.description,
		Subject:     t.subject,
		BodyHTML:    t.bodyHTML,
		BodyText:    t.bodyText,
		Variables:   vars,
		Status:      t.status.String(),
		Version:     t.version,
		IsSystem:    t.IsSystemTemplate(),
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
		CreatedAt:   t.createdAt.Time(),
		UpdatedAt:   t.updatedAt.Time(),
		ActivatedAt: nil, // TODO: add field if needed
		ArchivedAt:  nil, // TODO: add field if needed
	}
}

// FromSnapshot reconstitutes a template from a snapshot.
func FromSnapshot(s Snapshot) (*Template, error) {
	id, err := types.ParseID(s.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid template id: %w", err)
	}

	locale, err := ParseLocale(s.Locale)
	if err != nil {
		return nil, fmt.Errorf("invalid locale: %w", err)
	}

	vars := make(Variables, len(s.Variables))
	for i, v := range s.Variables {
		varType, err := ParseVariableType(v.Type)
		if err != nil {
			varType = VariableTypeString // default
		}
		vars[i] = Variable{
			Name:        v.Name,
			Type:        varType,
			Required:    v.Required,
			Default:     v.Default,
			Description: v.Description,
		}
	}

	var tenantID *string
	if s.TenantID != "" {
		tenantID = &s.TenantID
	}

	var createdBy, updatedBy *types.ID
	if s.CreatedBy != "" {
		parsed, err := types.ParseID(s.CreatedBy)
		if err == nil {
			createdBy = &parsed
		}
	}
	if s.UpdatedBy != "" {
		parsed, err := types.ParseID(s.UpdatedBy)
		if err == nil {
			updatedBy = &parsed
		}
	}

	return &Template{
		id:          id,
		tenantID:    tenantID,
		slug:        s.Slug,
		locale:      locale,
		name:        s.Name,
		description: s.Description,
		subject:     s.Subject,
		bodyHTML:    s.BodyHTML,
		bodyText:    s.BodyText,
		variables:   vars,
		status:      Status(s.Status),
		version:     s.Version,
		createdBy:   createdBy,
		updatedBy:   updatedBy,
		createdAt:   types.NewTimestamp(s.CreatedAt),
		updatedAt:   types.NewTimestamp(s.UpdatedAt),
		events:      make([]TemplateEvent, 0),
	}, nil
}
