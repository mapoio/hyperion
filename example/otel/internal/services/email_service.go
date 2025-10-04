package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// EmailService handles email operations
type EmailService struct {
	templateService *TemplateService
}

// NewEmailService creates a new EmailService instance
func NewEmailService(
	templateService *TemplateService,
) *EmailService {
	return &EmailService{
		templateService: templateService,
	}
}

// SendPaymentEmail sends payment confirmation email (Level 4)
func (s *EmailService) SendPaymentEmail(hctx hyperion.Context, userID, transactionID string, amount float64) (err error) {
	hctx, end := hctx.UseIntercept("EmailService", "SendPaymentEmail")
	defer end(&err)

	hctx.Logger().Info("sending payment email", "user_id", userID, "transaction_id", transactionID)

	// Render email template (Level 5)
	emailContent, err := s.templateService.RenderPaymentEmail(hctx, userID, transactionID, amount)
	if err != nil {
		hctx.Logger().Error("email template rendering failed", "error", err)
		return fmt.Errorf("template rendering failed: %w", err)
	}

	// Simulate email sending via SMTP
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	// Randomly fail 2% of the time
	if rand.Float64() < 0.02 {
		err := fmt.Errorf("SMTP server error")
		hctx.Logger().Error("email sending failed", "error", err)
		return err
	}

	hctx.Logger().Info("payment email sent successfully", "user_id", userID, "content_length", len(emailContent))

	return nil
}
