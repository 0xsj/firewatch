// internal/notifications/application/dto/mappers.go
package dto

import "github.com/0xsj/hexagonal-go/internal/notifications/domain"

// MapNotificationToDTO maps a domain notification to a DTO.
func MapNotificationToDTO(n *domain.Notification) *NotificationDTO {
	return &NotificationDTO{
		ID:            n.ID().String(),
		TenantID:      n.TenantID(),
		Channel:       n.Channel().String(),
		Recipient:     n.Recipient(),
		Subject:       n.Subject(),
		Status:        n.Status().String(),
		Attempts:      n.Attempts(),
		LastError:     n.LastError(),
		SentAt:        n.SentAt(),
		UserID:        n.UserID(),
		CorrelationID: n.CorrelationID(),
		EventType:     n.EventType(),
		CreatedAt:     n.CreatedAt(),
		UpdatedAt:     n.UpdatedAt(),
	}
}

// MapNotificationsToDTO maps a slice of domain notifications to DTOs.
func MapNotificationsToDTO(notifications []*domain.Notification) []*NotificationDTO {
	dtos := make([]*NotificationDTO, len(notifications))
	for i, n := range notifications {
		dtos[i] = MapNotificationToDTO(n)
	}
	return dtos
}
