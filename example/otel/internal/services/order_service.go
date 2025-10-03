package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mapoio/hyperion"
)

// OrderService handles order-related operations
type OrderService struct {
	paymentService   *PaymentService
	inventoryService *InventoryService
}

// NewOrderService creates a new OrderService instance
func NewOrderService(
	paymentService *PaymentService,
	inventoryService *InventoryService,
) *OrderService {
	return &OrderService{
		paymentService:   paymentService,
		inventoryService: inventoryService,
	}
}

// CreateOrder creates a new order (Level 1 - Entry point)
func (s *OrderService) CreateOrder(hctx hyperion.Context, userID, productID string, amount float64) (orderID string, err error) {
	// UseIntercept applies all registered interceptors (tracing, logging, etc.)
	// The TracingInterceptor automatically:
	// 1. Creates OpenTelemetry span for "OrderService.CreateOrder"
	// 2. Updates context with new span
	// 3. Records error on span if err != nil
	// 4. Ends span when function returns
	hctx, end := hctx.UseIntercept("OrderService", "CreateOrder")
	defer end(&err)

	// Logger automatically includes trace_id and span_id from context
	hctx.Logger().Info("creating order",
		"user_id", userID,
		"product_id", productID,
		"amount", amount,
	)

	// Simulate order validation
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	// Check inventory (Level 2)
	available, err := s.inventoryService.CheckStock(hctx, productID)
	if err != nil {
		hctx.Logger().Error("inventory check failed", "error", err)
		return "", fmt.Errorf("inventory check failed: %w", err)
	}

	if !available {
		hctx.Logger().Warn("product out of stock", "product_id", productID)
		return "", fmt.Errorf("product out of stock")
	}

	// Process payment (Level 2)
	transactionID, err := s.paymentService.ProcessPayment(hctx, userID, amount)
	if err != nil {
		hctx.Logger().Error("payment processing failed", "error", err)
		return "", fmt.Errorf("payment failed: %w", err)
	}

	// Reserve inventory (Level 2)
	if err := s.inventoryService.ReserveStock(hctx, productID, 1); err != nil {
		hctx.Logger().Error("inventory reservation failed", "error", err)
		// Rollback payment
		s.paymentService.RefundPayment(hctx, transactionID)
		return "", fmt.Errorf("inventory reservation failed: %w", err)
	}

	orderID = fmt.Sprintf("ORD-%d", time.Now().UnixNano())

	hctx.Logger().Info("order created successfully",
		"order_id", orderID,
		"transaction_id", transactionID,
	)

	return orderID, nil
}
