package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// PaymentService handles payment operations
type PaymentService struct {
	fraudService        *FraudDetectionService
	notificationService *NotificationService
}

// NewPaymentService creates a new PaymentService instance
func NewPaymentService(
	fraudService *FraudDetectionService,
	notificationService *NotificationService,
) *PaymentService {
	return &PaymentService{
		fraudService:        fraudService,
		notificationService: notificationService,
	}
}

// ProcessPayment processes a payment (Level 2)
func (s *PaymentService) ProcessPayment(hctx hyperion.Context, userID string, amount float64) (transactionID string, err error) {
	hctx, end := hctx.UseIntercept("PaymentService", "ProcessPayment")
	defer end(&err)

	hctx.Logger().Info("processing payment", "user_id", userID, "amount", amount)

	// Check for fraud (Level 3)
	isSafe, err := s.fraudService.CheckTransaction(hctx, userID, amount)
	if err != nil {
		hctx.Logger().Error("fraud check failed", "error", err)
		return "", fmt.Errorf("fraud check failed: %w", err)
	}

	if !isSafe {
		hctx.Logger().Warn("fraudulent transaction detected", "user_id", userID, "amount", amount)
		return "", fmt.Errorf("transaction flagged as fraudulent")
	}

	// Simulate payment gateway call
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)

	transactionID = fmt.Sprintf("TXN-%d", time.Now().UnixNano())

	// Send payment confirmation notification (Level 3)
	if err := s.notificationService.SendPaymentConfirmation(hctx, userID, transactionID, amount); err != nil {
		hctx.Logger().Warn("failed to send payment confirmation", "error", err)
		// Don't fail the payment if notification fails
	}

	hctx.Logger().Info("payment processed successfully", "transaction_id", transactionID)

	return transactionID, nil
}

// RefundPayment refunds a payment (Level 2)
func (s *PaymentService) RefundPayment(hctx hyperion.Context, transactionID string) (err error) {
	hctx, end := hctx.UseIntercept("PaymentService", "RefundPayment")
	defer end(&err)

	hctx.Logger().Info("refunding payment", "transaction_id", transactionID)

	// Simulate refund processing
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	hctx.Logger().Info("payment refunded successfully", "transaction_id", transactionID)

	return nil
}
