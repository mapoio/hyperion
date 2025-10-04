package services

import (

	"github.com/mapoio/hyperion"
)

// NotificationService handles notifications
type NotificationService struct {
	emailService *EmailService
	smsService   *SMSService
}

// NewNotificationService creates a new NotificationService instance
func NewNotificationService(
	emailService *EmailService,
	smsService *SMSService,
) *NotificationService {
	return &NotificationService{
		emailService: emailService,
		smsService:   smsService,
	}
}

// SendPaymentConfirmation sends payment confirmation (Level 3)
func (s *NotificationService) SendPaymentConfirmation(hctx hyperion.Context, userID, transactionID string, amount float64) (err error) {
	hctx, end := hctx.UseIntercept("NotificationService", "SendPaymentConfirmation")
	defer end(&err)

	hctx.Logger().Info("sending payment confirmation", "user_id", userID, "transaction_id", transactionID)

	// Send email notification (Level 4)
	if err := s.emailService.SendPaymentEmail(hctx, userID, transactionID, amount); err != nil {
		hctx.Logger().Error("email notification failed", "error", err)
		// Continue to try SMS
	}

	// Send SMS notification (Level 4)
	if err := s.smsService.SendPaymentSMS(hctx, userID, transactionID, amount); err != nil {
		hctx.Logger().Error("SMS notification failed", "error", err)
		return err
	}

	hctx.Logger().Info("payment confirmation sent successfully", "user_id", userID)

	return nil
}
