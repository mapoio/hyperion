package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// SMSService handles SMS operations
type SMSService struct {
	templateService *TemplateService
}

// NewSMSService creates a new SMSService instance
func NewSMSService(
	templateService *TemplateService,
) *SMSService {
	return &SMSService{
		templateService: templateService,
	}
}

// SendPaymentSMS sends payment confirmation SMS (Level 4)
func (s *SMSService) SendPaymentSMS(hctx hyperion.Context, userID, transactionID string, amount float64) (err error) {
	hctx, end := hctx.UseIntercept("SMSService", "SendPaymentSMS")
	defer end(&err)

	hctx.Logger().Info("sending payment SMS", "user_id", userID, "transaction_id", transactionID)

	// Render SMS template (Level 5)
	smsContent, err := s.templateService.RenderPaymentSMS(hctx, userID, transactionID, amount)
	if err != nil {
		hctx.Logger().Error("SMS template rendering failed", "error", err)
		return fmt.Errorf("template rendering failed: %w", err)
	}

	// Simulate SMS gateway call
	time.Sleep(time.Duration(rand.Intn(15)) * time.Millisecond)

	// Randomly fail 3% of the time
	if rand.Float64() < 0.03 {
		err := fmt.Errorf("SMS gateway error")
		hctx.Logger().Error("SMS sending failed", "error", err)
		return err
	}

	hctx.Logger().Info("payment SMS sent successfully", "user_id", userID, "content_length", len(smsContent))

	return nil
}
