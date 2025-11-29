package tenant

import (
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/types"
)

// Tenant is the aggregate root for tenant management.
// Encapsulates all business logic related to tenant lifecycle.
//
// Design principles:
//   - Aggregate root enforces invariants
//   - No public setters - use methods that express intent
//   - Collects domain events for publishing
//   - Self-contained, no dependencies on other domain modules
type Tenant struct {
	// Identity
	id   types.ID
	slug Slug
	name string

	// Subscription
	plan   Plan
	status Status

	// Configuration
	settings Settings

	// Ownership
	ownerID    types.ID
	ownerEmail string

	// Billing integration (optional)
	billingID   *string
	trialEndsAt *types.Timestamp

	// Timestamps
	createdAt types.Timestamp
	updatedAt types.Timestamp

	// Domain events (uncommitted)
	events []Event

	// Aggregate version (for optimistic locking)
	version int
}

// ============================================================================
// Aggregate Getters (Read-only access)
// ============================================================================

func (t *Tenant) ID() types.ID                  { return t.id }
func (t *Tenant) Slug() Slug                    { return t.slug }
func (t *Tenant) Name() string                  { return t.name }
func (t *Tenant) Plan() Plan                    { return t.plan }
func (t *Tenant) Status() Status                { return t.status }
func (t *Tenant) Settings() Settings            { return t.settings }
func (t *Tenant) OwnerID() types.ID             { return t.ownerID }
func (t *Tenant) OwnerEmail() string            { return t.ownerEmail }
func (t *Tenant) BillingID() *string            { return t.billingID }
func (t *Tenant) TrialEndsAt() *types.Timestamp { return t.trialEndsAt }
func (t *Tenant) CreatedAt() types.Timestamp    { return t.createdAt }
func (t *Tenant) UpdatedAt() types.Timestamp    { return t.updatedAt }
func (t *Tenant) Version() int                  { return t.version }

// IsActive returns true if the tenant can operate normally.
func (t *Tenant) IsActive() bool {
	return t.status.IsActive()
}

// IsAccessible returns true if the tenant's data can be accessed.
func (t *Tenant) IsAccessible() bool {
	return t.status.IsAccessible()
}

// IsPaidPlan returns true if the tenant is on a paid plan.
func (t *Tenant) IsPaidPlan() bool {
	return t.plan.IsPaid()
}

// ============================================================================
// Factory Methods (Aggregate Creation)
// ============================================================================

// Create creates a new tenant.
// Emits TenantCreated event.
func Create(
	id types.ID,
	slug Slug,
	name string,
	plan Plan,
	ownerID types.ID,
	ownerEmail string,
	createdBy string,
) (*Tenant, error) {
	const op = "tenant.Create"

	// Validate inputs
	if id.IsEmpty() {
		return nil, fmt.Errorf("%s: tenant id is required", op)
	}
	if slug.IsEmpty() {
		return nil, fmt.Errorf("%s: slug is required", op)
	}
	if name == "" {
		return nil, ErrTenantNameInvalid(op, name, "name is required")
	}
	if len(name) > 255 {
		return nil, ErrTenantNameInvalid(op, name, "name must be at most 255 characters")
	}
	if err := plan.Validate(); err != nil {
		return nil, ErrPlanInvalid(op, plan.String())
	}
	if ownerID.IsEmpty() {
		return nil, fmt.Errorf("%s: owner id is required", op)
	}
	if ownerEmail == "" {
		return nil, fmt.Errorf("%s: owner email is required", op)
	}

	now := types.Now()

	tenant := &Tenant{
		id:          id,
		slug:        slug,
		name:        name,
		plan:        plan,
		status:      StatusActive,
		settings:    NewSettings(),
		ownerID:     ownerID,
		ownerEmail:  ownerEmail,
		billingID:   nil,
		trialEndsAt: nil,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]Event, 0),
		version:     1,
	}

	// Emit domain event
	tenant.addEvent(NewTenantCreated(
		id, slug, name, plan, ownerID, ownerEmail, createdBy,
	))

	return tenant, nil
}

// CreateWithTrial creates a new tenant with a trial period.
// Emits TenantCreated event.
func CreateWithTrial(
	id types.ID,
	slug Slug,
	name string,
	plan Plan,
	ownerID types.ID,
	ownerEmail string,
	trialEndsAt types.Timestamp,
	createdBy string,
) (*Tenant, error) {
	const op = "tenant.CreateWithTrial"

	tenant, err := Create(id, slug, name, plan, ownerID, ownerEmail, createdBy)
	if err != nil {
		return nil, err
	}

	// Override status to trialing
	tenant.status = StatusTrialing
	tenant.trialEndsAt = &trialEndsAt

	// Clear and re-emit event with correct status
	tenant.events = make([]Event, 0)
	tenant.addEvent(NewTenantCreated(
		id, slug, name, plan, ownerID, ownerEmail, createdBy,
	))

	return tenant, nil
}

// Reconstitute recreates a tenant from stored state (used by repository).
// Does NOT emit events - only for loading from database.
func Reconstitute(
	id types.ID,
	slug Slug,
	name string,
	plan Plan,
	status Status,
	settings Settings,
	ownerID types.ID,
	ownerEmail string,
	billingID *string,
	trialEndsAt *types.Timestamp,
	createdAt types.Timestamp,
	updatedAt types.Timestamp,
	version int,
) *Tenant {
	return &Tenant{
		id:          id,
		slug:        slug,
		name:        name,
		plan:        plan,
		status:      status,
		settings:    settings,
		ownerID:     ownerID,
		ownerEmail:  ownerEmail,
		billingID:   billingID,
		trialEndsAt: trialEndsAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]Event, 0),
		version:     version,
	}
}

// ============================================================================
// Tenant Updates
// ============================================================================

// Update updates the tenant's name.
// Emits TenantUpdated event.
func (t *Tenant) Update(name string, updatedBy string) error {
	const op = "tenant.Update"

	if t.status == StatusDeleted {
		return ErrTenantDeleted(op, t.id.String())
	}

	var updatedFields []string

	if name != "" && name != t.name {
		if len(name) > 255 {
			return ErrTenantNameInvalid(op, name, "name must be at most 255 characters")
		}
		t.name = name
		updatedFields = append(updatedFields, "name")
	}

	if len(updatedFields) == 0 {
		return nil // No changes
	}

	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantUpdated(t.id, t.ownerEmail, updatedFields, updatedBy))

	return nil
}

// UpdateSettings updates the tenant's settings.
// Emits TenantSettingsUpdated event.
func (t *Tenant) UpdateSettings(settings Settings, updatedBy string) error {
	const op = "tenant.UpdateSettings"

	if t.status == StatusDeleted {
		return ErrTenantDeleted(op, t.id.String())
	}

	// Determine changed keys
	changedKeys := make([]string, 0)
	newMap := settings.ToMap()
	oldMap := t.settings.ToMap()

	for k, v := range newMap {
		if oldVal, exists := oldMap[k]; !exists || oldVal != v {
			changedKeys = append(changedKeys, k)
		}
	}
	for k := range oldMap {
		if _, exists := newMap[k]; !exists {
			changedKeys = append(changedKeys, k)
		}
	}

	if len(changedKeys) == 0 {
		return nil // No changes
	}

	t.settings = settings
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantSettingsUpdated(t.id, t.ownerEmail, changedKeys, updatedBy))

	return nil
}

// SetBillingID sets the external billing system ID.
func (t *Tenant) SetBillingID(billingID string, updatedBy string) error {
	const op = "tenant.SetBillingID"

	if t.status == StatusDeleted {
		return ErrTenantDeleted(op, t.id.String())
	}

	t.billingID = &billingID
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantUpdated(t.id, t.ownerEmail, []string{"billing_id"}, updatedBy))

	return nil
}

// SetOwnerEmail updates the owner email (used when owner changes or email is updated).
func (t *Tenant) SetOwnerEmail(ownerEmail string) {
	t.ownerEmail = ownerEmail
	t.updatedAt = types.Now()
}

// ============================================================================
// Plan Management
// ============================================================================

// ChangePlan changes the tenant's subscription plan.
// Emits TenantPlanChanged event.
func (t *Tenant) ChangePlan(newPlan Plan, changedBy string, reason string) error {
	const op = "tenant.ChangePlan"

	if t.status == StatusDeleted {
		return ErrTenantDeleted(op, t.id.String())
	}

	if err := newPlan.Validate(); err != nil {
		return ErrPlanInvalid(op, newPlan.String())
	}

	if t.plan == newPlan {
		return nil // No change
	}

	oldPlan := t.plan
	t.plan = newPlan
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantPlanChanged(t.id, t.ownerEmail, oldPlan, newPlan, changedBy, reason))

	return nil
}

// ============================================================================
// Status Management
// ============================================================================

// Suspend suspends the tenant.
// Emits TenantSuspended event.
func (t *Tenant) Suspend(reason string, suspendedBy string) error {
	const op = "tenant.Suspend"

	if err := t.status.CanTransitionTo(StatusSuspended); err != nil {
		return ErrInvalidStatusChange(op, t.status.String(), StatusSuspended.String())
	}

	t.status = StatusSuspended
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantSuspended(t.id, t.ownerEmail, reason, suspendedBy))

	return nil
}

// Reactivate reactivates a suspended tenant.
// Emits TenantReactivated event.
func (t *Tenant) Reactivate(reactivatedBy string) error {
	const op = "tenant.Reactivate"

	if t.status != StatusSuspended {
		return ErrInvalidStatusChange(op, t.status.String(), StatusActive.String())
	}

	t.status = StatusActive
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantReactivated(t.id, t.ownerEmail, reactivatedBy))

	return nil
}

// Cancel cancels the tenant's subscription.
// Emits TenantDeleted event with reason.
func (t *Tenant) Cancel(reason string, cancelledBy string) error {
	const op = "tenant.Cancel"

	if err := t.status.CanTransitionTo(StatusCancelled); err != nil {
		return ErrInvalidStatusChange(op, t.status.String(), StatusCancelled.String())
	}

	t.status = StatusCancelled
	t.updatedAt = types.Now()

	return nil
}

// Delete soft-deletes the tenant.
// Emits TenantDeleted event.
func (t *Tenant) Delete(reason string, deletedBy string) error {
	const op = "tenant.Delete"

	if err := t.status.CanTransitionTo(StatusDeleted); err != nil {
		return ErrInvalidStatusChange(op, t.status.String(), StatusDeleted.String())
	}

	t.status = StatusDeleted
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantDeleted(t.id, t.ownerEmail, reason, deletedBy))

	return nil
}

// Activate activates a trialing tenant.
// Transitions from trialing to active status.
func (t *Tenant) Activate(activatedBy string) error {
	const op = "tenant.Activate"

	if t.status != StatusTrialing {
		return ErrInvalidStatusChange(op, t.status.String(), StatusActive.String())
	}

	t.status = StatusActive
	t.trialEndsAt = nil
	t.updatedAt = types.Now()

	// Emit event
	t.addEvent(NewTenantReactivated(t.id, t.ownerEmail, activatedBy))

	return nil
}

// ============================================================================
// Event Management
// ============================================================================

// Events returns all uncommitted domain events.
func (t *Tenant) Events() []Event {
	return t.events
}

// ClearEvents clears all uncommitted events.
// Called after events are published.
func (t *Tenant) ClearEvents() {
	t.events = make([]Event, 0)
}

// addEvent adds a domain event to the uncommitted events list.
func (t *Tenant) addEvent(event Event) {
	t.events = append(t.events, event)
}

// ============================================================================
// Version Management (Optimistic Locking)
// ============================================================================

// IncrementVersion increments the aggregate version.
// Used for optimistic locking in the repository.
func (t *Tenant) IncrementVersion() {
	t.version++
	t.updatedAt = types.Now()
}
