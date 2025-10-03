package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// TemplateService handles template rendering
type TemplateService struct {
	storageService *StorageService
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService(
	storageService *StorageService,
) *TemplateService {
	return &TemplateService{
		storageService: storageService,
	}
}

// RenderPaymentEmail renders payment email template (Level 5)
func (s *TemplateService) RenderPaymentEmail(hctx hyperion.Context, userID, transactionID string, amount float64) (content string, err error) {
	hctx, end := hctx.UseIntercept("TemplateService", "RenderPaymentEmail")
	defer end(&err)

	hctx.Logger().Info("rendering payment email template", "user_id", userID, "transaction_id", transactionID)

	// Load template from storage (Level 6)
	template, err := s.storageService.LoadTemplate(hctx, "payment_email")
	if err != nil {
		hctx.Logger().Error("template loading failed", "error", err)
		return "", fmt.Errorf("template loading failed: %w", err)
	}

	// Simulate template rendering
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	content = fmt.Sprintf("%s - Payment of $%.2f processed. Transaction: %s", template, amount, transactionID)

	hctx.Logger().Info("payment email template rendered", "content_length", len(content))

	return content, nil
}

// RenderPaymentSMS renders payment SMS template (Level 5)
func (s *TemplateService) RenderPaymentSMS(hctx hyperion.Context, userID, transactionID string, amount float64) (content string, err error) {
	hctx, end := hctx.UseIntercept("TemplateService", "RenderPaymentSMS")
	defer end(&err)

	hctx.Logger().Info("rendering payment SMS template", "user_id", userID, "transaction_id", transactionID)

	// Load template from storage (Level 6)
	template, err := s.storageService.LoadTemplate(hctx, "payment_sms")
	if err != nil {
		hctx.Logger().Error("template loading failed", "error", err)
		return "", fmt.Errorf("template loading failed: %w", err)
	}

	// Simulate template rendering
	time.Sleep(time.Duration(rand.Intn(8)) * time.Millisecond)

	content = fmt.Sprintf("%s - $%.2f paid. Ref: %s", template, amount, transactionID)

	hctx.Logger().Info("payment SMS template rendered", "content_length", len(content))

	return content, nil
}
