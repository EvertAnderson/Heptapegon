package service

import (
	"context"
	"fmt"
	"log"

	"github.com/heptapegon/localpickup/internal/domain"
	"github.com/heptapegon/localpickup/pkg/fcm"
)

type NotificationService struct {
	fcm *fcm.Client
}

func NewNotificationService(fcmClient *fcm.Client) *NotificationService {
	return &NotificationService{fcm: fcmClient}
}

// SendNewOrderNotification pushes an FCM alert to the business when a payment
// is confirmed. Called asynchronously from OrderService.Create.
func (s *NotificationService) SendNewOrderNotification(ctx context.Context, fcmToken string, o *domain.Order) {
	if fcmToken == "" {
		return
	}

	shortID := o.ID.String()[:8]
	msg := &fcm.Message{
		Token: fcmToken,
		Notification: &fcm.Notification{
			Title: "Nuevo Pedido Recibido",
			Body:  fmt.Sprintf("Pedido #%s por $%.2f — ¡prepáralo!", shortID, o.TotalAmount),
		},
		Data: map[string]string{
			"type":     "new_order",
			"order_id": o.ID.String(),
		},
	}

	if err := s.fcm.Send(ctx, msg); err != nil {
		log.Printf("notification: FCM send failed for order %s: %v", o.ID, err)
	}
}

// SendOrderReadyNotification notifies the customer that the order is ready for pickup.
func (s *NotificationService) SendOrderReadyNotification(ctx context.Context, customerFCMToken string, o *domain.Order) {
	if customerFCMToken == "" {
		return
	}

	msg := &fcm.Message{
		Token: customerFCMToken,
		Notification: &fcm.Notification{
			Title: "¡Tu pedido está listo!",
			Body:  fmt.Sprintf("Pedido #%s está listo. Muestra tu PIN al retirar.", o.ID.String()[:8]),
		},
		Data: map[string]string{
			"type":     "order_ready",
			"order_id": o.ID.String(),
		},
	}

	if err := s.fcm.Send(ctx, msg); err != nil {
		log.Printf("notification: FCM send failed: %v", err)
	}
}
