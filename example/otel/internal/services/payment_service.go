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
	// Metrics
	paymentCounter  hyperion.Counter   // Total payments processed
	paymentAmount   hyperion.Histogram // Payment amount distribution
	paymentDuration hyperion.Histogram // Payment processing duration
	refundCounter   hyperion.Counter   // Total refunds processed
	refundDuration  hyperion.Histogram // Refund processing duration
}

// NewPaymentService creates a new PaymentService instance
func NewPaymentService(
	fraudService *FraudDetectionService,
	notificationService *NotificationService,
	meter hyperion.Meter,
) *PaymentService {
	return &PaymentService{
		fraudService:        fraudService,
		notificationService: notificationService,
		// Initialize metrics
		paymentCounter: meter.Counter("payment.processed.total",
			hyperion.WithMetricDescription("Total number of payments processed"),
			hyperion.WithMetricUnit("1"),
		),
		paymentAmount: meter.Histogram("payment.amount",
			hyperion.WithMetricDescription("Payment amount distribution"),
			hyperion.WithMetricUnit("USD"),
		),
		paymentDuration: meter.Histogram("payment.processing.duration",
			hyperion.WithMetricDescription("Payment processing duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
		refundCounter: meter.Counter("payment.refund.total",
			hyperion.WithMetricDescription("Total number of refunds processed"),
			hyperion.WithMetricUnit("1"),
		),
		refundDuration: meter.Histogram("payment.refund.duration",
			hyperion.WithMetricDescription("Refund processing duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
	}
}

// ProcessPayment processes a payment (Level 2)
func (s *PaymentService) ProcessPayment(hctx hyperion.Context, userID string, amount float64) (transactionID string, err error) {
	// Track payment processing start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.paymentDuration.Record(hctx, duration,
			hyperion.String("service", "payment"),
			hyperion.String("operation", "process"),
			hyperion.String("status", status),
		)

		// Record payment metrics on success
		if err == nil {
			s.paymentCounter.Add(hctx, 1,
				hyperion.String("service", "payment"),
				hyperion.String("status", "success"),
			)
			s.paymentAmount.Record(hctx, amount,
				hyperion.String("service", "payment"),
			)
		} else {
			s.paymentCounter.Add(hctx, 1,
				hyperion.String("service", "payment"),
				hyperion.String("status", "error"),
			)
		}
	}()

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
	// Track refund processing start time
	startTime := time.Now()

	defer func() {
		// Record processing duration
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.refundDuration.Record(hctx, duration,
			hyperion.String("service", "payment"),
			hyperion.String("operation", "refund"),
			hyperion.String("status", status),
		)

		// Record refund counter
		s.refundCounter.Add(hctx, 1,
			hyperion.String("service", "payment"),
			hyperion.String("status", status),
		)
	}()

	hctx, end := hctx.UseIntercept("PaymentService", "RefundPayment")
	defer end(&err)

	hctx.Logger().Info("refunding payment", "transaction_id", transactionID)

	// Simulate refund processing
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	hctx.Logger().Info("payment refunded successfully", "transaction_id", transactionID)

	return nil
}
